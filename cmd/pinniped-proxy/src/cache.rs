// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use std::{collections::HashMap, hash::Hash, sync::RwLock};

/// An unlimited cache that can be RwLocked across threads.
///
/// Importantly, checking the cache does not require a write-lock
/// (unlike the [`Cached` trait's `cache_get`](https://github.com/jaemk/cached/blob/f5911dc3fbc03e1db9f87192eb854fac2ee6ac98/src/lib.rs#L203))
#[derive(Default)]
struct LockableCache<K, V>(RwLock<HashMap<K, V>>);

impl<K, V> LockableCache<K, V> {
    fn get(&self, k: &K) -> Option<V>
    where
        K: Eq + Hash,
        V: Eq + Hash + Clone,
    {
        (*self.0.read().unwrap())
            .get(k)
            .and_then(|v| Some((*v).clone()))
    }

    // insert is called in PrunableCache via a write lock guard, but compiler
    // doesn't see this, apparently.
    #[allow(dead_code)]
    fn insert(&self, k: K, v: V)
    where
        K: Eq + Hash,
        V: Eq + Hash,
    {
        self.0.write().unwrap().insert(k, v);
    }

    #[cfg(test)]
    fn len(&self) -> usize {
        (*self.0.read().unwrap()).len()
    }
}

/// A cache that additionally prunes itself whenever it holds the write-lock.
pub struct PruningCache<K, V> {
    cache: LockableCache<K, V>,
    prune_fn: fn(&(K, V)) -> bool,
}

impl<K, V> PruningCache<K, V> {
    pub fn new(f: fn(&(K, V)) -> bool) -> PruningCache<K, V>
    where
        K: Default,
        V: Default,
    {
        PruningCache {
            cache: LockableCache::default(),
            prune_fn: f,
        }
    }

    pub fn get(&self, k: &K) -> Option<V>
    where
        K: Eq + Hash + Clone,
        V: Eq + Hash + Clone,
    {
        // Only return the value from the cache if it should not have been
        // pruned.
        self.cache
            .get(k)
            .and_then(|v| (self.prune_fn)(&(k.clone(), v.clone())).then_some(v))
    }

    // Prunes the cache while holding the write-lock during an insert.
    pub fn insert(&self, k: K, v: V)
    where
        K: Eq + Hash + Clone,
        V: Eq + Hash + Clone,
    {
        let mut write_guard = self.cache.0.write().unwrap();
        write_guard.insert(k, v);
        // Replace the cache with one where items are pruned.
        let cache = std::mem::take(&mut *write_guard);
        *write_guard = cache.into_iter().filter(self.prune_fn).collect();
    }

    #[cfg(test)]
    pub fn len(&self) -> usize {
        self.cache.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lockable_cache_get() {
        let c = LockableCache::default();

        c.insert(1, 5);
        c.insert(2, 3);

        assert_eq!(c.get(&1), Some(5));
        assert_eq!(c.get(&2), Some(3));
        assert_eq!(c.len(), 2);
    }

    #[test]
    fn test_lockable_cache_overwrite() {
        let c = LockableCache::default();

        c.insert(1, 5);
        assert_eq!(c.get(&1), Some(5));

        c.insert(1, 6);
        assert_eq!(c.get(&1), Some(6));
        assert_eq!(c.len(), 1);
    }

    #[test]
    fn test_pruning_cache_get() {
        // Create a cache that prunes all odd numbers (keeps evens only).
        let c = PruningCache::new(|(_k, v)| *v % 2 == 0);

        c.insert(1, 1);
        c.insert(2, 2);
        c.insert(3, 3);
        c.insert(4, 4);

        assert_eq!(c.get(&1), None);
        assert_eq!(c.get(&2), Some(2));
        assert_eq!(c.get(&3), None);
        assert_eq!(c.get(&4), Some(4));
        assert_eq!(c.len(), 2);
    }
}

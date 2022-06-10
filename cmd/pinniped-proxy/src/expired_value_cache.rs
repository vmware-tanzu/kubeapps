// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// NOTE: This may be better in the upstream cached repository. I've created
// an issue to check with the author, otherwise it can stay here or I may
// spin it off to a separate crate.
// https://github.com/jaemk/cached/issues/115
use std::collections::HashMap;
use std::hash::Hash;

use cached::Cached;

pub trait CanExpire {
    fn is_expired(&self) -> bool;
}

// Implement our ExpiredValueCache which uses the returned values own
// trait to determine whether it has expired, rather than a set timestamp.
struct ExpiredValueCache<K: Hash + Eq, V: CanExpire> {
    store: HashMap<K, V>,
    capacity: usize,
    hits: u64,
    misses: u64,
}
impl<K: Hash + Eq, V: CanExpire> ExpiredValueCache<K, V> {
    pub fn with_capacity(size: usize) -> ExpiredValueCache<K, V> {
        ExpiredValueCache {
            store: HashMap::with_capacity(size),
            capacity: size,
            hits: 0,
            misses: 0,
        }
    }
}

impl<K: Hash + Eq, V: CanExpire> Cached<K, V> for ExpiredValueCache<K, V> {
    fn cache_get(&mut self, k: &K) -> Option<&V> {
        let optv = self.store.get(k);
        // If it has expired, delete the cached entry and return none.
        match optv {
            Some(v) => {
                if v.is_expired() {
                    self.cache_remove(k);
                    self.misses += 1;
                    return None;
                }
            },
            None => {
                self.misses += 1;
                return None
            },
        }
        // We cannot simply return optv here because the ownership of the value
        // moved to the self.cache_remove(k), so optv cannot be referenced after
        // that line.
        self.hits += 1;
        self.store.get(k)
    }

    fn cache_get_mut(&mut self, k: &K) -> Option<&mut V> {
        let optv = self.store.get(k);
        // If it has expired, delete the cached entry and return none.
        match optv {
            Some(v) => {
                if v.is_expired() {
                    self.cache_remove(k);
                    self.misses += 1;
                    return None;
                }
            },
            None => {
                self.misses += 1;
                return None
            }
        }
        self.hits += 1;
        self.store.get_mut(k)
    }

    fn cache_get_or_set_with<F: FnOnce() -> V>(&mut self, k: K, f: F) -> &mut V {
        self.store.entry(k).or_insert_with(f)
    }
    fn cache_set(&mut self, k: K, v: V) -> Option<V> {
        self.store.insert(k, v)
    }
    fn cache_remove(&mut self, k: &K) -> Option<V> {
        self.store.remove(k)
    }
    fn cache_clear(&mut self) {
        self.store.clear();
    }
    fn cache_reset(&mut self) {
        self.store = HashMap::with_capacity(self.capacity);
    }
    fn cache_size(&self) -> usize {
        self.store.len()
    }
    fn cache_hits(&self) -> Option<u64> {
        Some(self.hits)
    }
    fn cache_misses(&self) -> Option<u64> {
        Some(self.misses)
    }
    fn cache_reset_metrics(&mut self) {
        self.misses = 0;
        self.hits = 0;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serial_test::serial;

    use cached::proc_macro::cached;

    #[derive(Clone)]
    pub struct NewsArticle {
        slug: String,
        is_expired: bool,
    }

    impl CanExpire for NewsArticle {
        fn is_expired(&self) -> bool {
            self.is_expired
        }
    }

    const EXPIRED_SLUG: &str = "expired_slug";
    const UNEXPIRED_SLUG: &str = "unexpired_slug";

    #[cached(
        type = "ExpiredValueCache<String, NewsArticle>",
        create = "{ ExpiredValueCache::with_capacity(3) }",
        result = true,
    )]
    fn fetch_article(slug: String) -> Result<NewsArticle, ()> {
        match slug.as_str() {
            EXPIRED_SLUG => Ok(NewsArticle {
                slug: String::from(EXPIRED_SLUG),
                is_expired: true,
            }),
            UNEXPIRED_SLUG => Ok(NewsArticle {
                slug: String::from(UNEXPIRED_SLUG),
                is_expired: false,
            }),
            _ => Err(())
        }
    }

    #[test]
    #[serial(cachetest)]
    fn test_expired_article_returned_with_miss() {
        {
            let mut cache = FETCH_ARTICLE.lock().unwrap();
            cache.cache_reset();
            cache.cache_reset_metrics();
        }
        let expired_article = fetch_article(EXPIRED_SLUG.to_string());

        assert!(expired_article.is_ok());
        assert_eq!(EXPIRED_SLUG, expired_article.unwrap().slug.as_str());

        // The article was fetched due to a cache miss and the result cached.
        {
            let cache = FETCH_ARTICLE.lock().unwrap();
            assert_eq!(1, cache.cache_size());
            assert_eq!(cache.cache_hits(), Some(0));
            assert_eq!(cache.cache_misses(), Some(1));
        }

        let _ = fetch_article(EXPIRED_SLUG.to_string());

        // The article was fetched again as it had expired.
        {
            let cache = FETCH_ARTICLE.lock().unwrap();
            assert_eq!(1, cache.cache_size());
            assert_eq!(cache.cache_hits(), Some(0));
            assert_eq!(cache.cache_misses(), Some(2));
        }
    }

    #[test]
    #[serial(cachetest)]
    fn test_unexpired_article_returned_with_hit() {
        {
            let mut cache = FETCH_ARTICLE.lock().unwrap();
            cache.cache_reset();
            cache.cache_reset_metrics();
        }
        let unexpired_article = fetch_article(UNEXPIRED_SLUG.to_string());

        assert!(unexpired_article.is_ok());
        assert_eq!(UNEXPIRED_SLUG, unexpired_article.unwrap().slug.as_str());

        // The article was fetched due to a cache miss and the result cached.
        {
            let cache = FETCH_ARTICLE.lock().unwrap();
            assert_eq!(1, cache.cache_size());
            assert_eq!(cache.cache_hits(), Some(0));
            assert_eq!(cache.cache_misses(), Some(1));
        }

        let cached_article = fetch_article(UNEXPIRED_SLUG.to_string());
        assert!(cached_article.is_ok());
        assert_eq!(UNEXPIRED_SLUG, cached_article.unwrap().slug.as_str());

        // The article was not fetched but returned as a hit from the cache.
        {
            let cache = FETCH_ARTICLE.lock().unwrap();
            assert_eq!(1, cache.cache_size());
            assert_eq!(cache.cache_hits(), Some(1));
            assert_eq!(cache.cache_misses(), Some(1));
        }
    }
}

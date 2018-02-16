These icons are generated using svgr using SVGs from the open-iconic project.

In order to add a new icon, install svgr:

```
yarn global install svgr
```

Find the icon SVG file you want to use and edit it to ensure it sets `fill="currentColor"` in the path element.

Then run (replacing with the icon you want to use):

```
svgr --no-semi --icon node_modules/open-iconic/svg/cog.svg > src/icons/Cog.tsx
```

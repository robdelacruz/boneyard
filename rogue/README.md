# rogue
A roguelike online playground in HTML5 Canvas and TypeScript.

2D virtual environment where online actors can view world items and chat with other actors.

## Status
- Define world tiles using plain text format.
- Define building and items in world.
- Display tiles in a grid.
- Use offscreen canvas buffers.

### Modules (so far):
- TileSheet
- Sprite
- Grid
- OffBuf
- Kb
- World

## How to build and run
```
$ npm run build-server        // build server code
$ npm run build-client        // build client code
$ npm run deploy-client       // deploy client code and website
$ npm run deploy-client-lib   // deploy client thirdparty libs
$ npm start                   // start node service

```

After building and deploying client, website client files (html, js) will be available in __public__ directory.

After building server, the runnable node.js scripts will be available in __dist/server__ directory.

## Screenshot
![screenshot](https://robdelacruz.github.io/images/rogue_screenshot.png "Screenshot")

## Demo
None yet.


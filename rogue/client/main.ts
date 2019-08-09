import {Pos,Log,LoadImages,ImgMap} from "./util";
import {TileSheet} from "./tile";
import {OffBuf,NewOffBufCanvas,NewOffBuf} from "./offbuf";
import {World} from "./world";
import io from "socket.io-client";
import * as rmlib from "./room";

let _currentRoom = rmlib.BlankRoom();

function main() {
  Log("Requesting room settings...");

  const socket = io();

  const req = <rmlib.RoomInfoReq>{};
  req.id = "great-hall";
  req.what = ["settings"];
  socket.emit("RoomInfoReq", req);

  // Client receives new <RoomInfo>
  // Copy roominfo fields to current room.
  socket.on("RoomInfo", function(rminfo:rmlib.RoomInfo) {
    // Changed room
    if (rminfo.id != _currentRoom.id) {
      _currentRoom = rmlib.BlankRoom();
    }

    rmlib.SetRoomInfo(_currentRoom, rminfo);
    console.log(_currentRoom);
  });

  const imageFiles = [
    "images/tiles_rpg.png",
    "images/tiles_indoor.png",
    "images/tiles_chars.png",
  ];
  LoadImages(imageFiles, function(imgs:ImgMap) {
    assetsLoaded(imgs);
  });
}
main();

function randInt(n:number):number {
  return Math.floor(Math.random() * n);
}

function assetsLoaded(imgs:ImgMap) {
  Log("assets loaded.");

  const imgChars = imgs["images/tiles_chars.png"];
  const tsCharsAliases = {
    "@1": [6,1],
    "@2": [7,1],
    "@3": [8,1],
  };
  const tsChars = new TileSheet(imgChars, 16,16,1, tsCharsAliases);

  const imgRpg = imgs["images/tiles_rpg.png"];
  const tsRpgAliases = {
    "W1": [25,0],
    "W2": [25,1],
    "W3": [25,2],
    "W4": [26,0],
    "W5": [26,1],
    "W6": [26,2],
    "W7": [27,0],
    "W8": [27,1],
    "W9": [27,2],
    "F1": [0,8],
    "F2": [0,7],
    "D1": [0,32],
    "D2": [0,33],
  };
  const tsRpg = new TileSheet(imgRpg, 16,16,1, tsRpgAliases);

  const imgIndoor = imgs["images/tiles_indoor.png"];
  const tsIndoorAliases = {
    "Ta": [0,3],
    "Tb": [0,4],
    "T2": [3,4],
    "C1": [4,2],
    "C2": [4,3],
    "C3": [4,0],
    "C4": [4,1],
  };
  const tsIndoor = new TileSheet(imgIndoor, 16,16,1, tsIndoorAliases);

  start(tsRpg, tsIndoor, tsChars);
}

function start(tsRpg:TileSheet, tsIndoor:TileSheet, tsChars:TileSheet) {
  Log("start");

  const cv = <HTMLCanvasElement>document.getElementById("canvas1");
  const ctx:CanvasRenderingContext2D = cv.getContext("2d")!;

/*
0  1  2  3  4  5  6  7  8  9  10 11 12 13 14  */
  const terrainLayer0 = `
W5 W8 W8 W8 W8 W8 W5 W5 W5 W5 W5 W5 W5 W5 W5
W6 F1 F1 F1 F1 F1 W4 W5 W5 W5 W5 W5 W5 W5 W5
W6 F1 F1 F1 F1 F1 W4 W8 W8 W8 W8 W5 W8 W8 W8
W6 F1 W1 W2 W3 F1 W4 F2 F2 F2 F2 W5 F2 F2 W4
W6 F1 W7 W8 W9 F1 W4 F2 F2 F2 F2 W5 F2 F2 W4
W6 F1 F1 F1 F1 F1 W4 F2 F2 F2 F2 W5 F2 F2 W4
W6 F1 F1 F1 F1 F1 W4 W8 W7 W5 W5 W5 W6 W8 W5
W6 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 W2
W6 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 F1 W2
W5 W2 W2 W2 W2 W2 W2 W2 W2 W2 W2 W2 W2 W2 W2
`;

  const terrainLayer1 = `
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- D1 -- -- -- -- -- D2 --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
`;

  const terrainLayer2 = `
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- C1 T2 C2 -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
-- -- C3 C3 -- -- -- -- -- -- -- -- -- -- --
-- -- Ta Tb -- -- -- -- -- -- -- -- -- -- --
-- -- C4 C4 -- -- -- -- -- -- -- -- -- -- --
-- -- -- -- -- -- -- -- -- -- -- -- -- -- --
`;

  const zones = `
## ## ## ## ## ## ## ## ## ## ## ## ## ## ##
## .. .. .. .. .. ## ## ## ## ## ## ## ## ##
## .. .. .. .. .. ## ## ## ## ## ## ## ## ##
## .. ## ## ## .. ## .. .. .. .. ## .. .. ##
## .. ## ## ## .. ## .. .. .. .. ## .. .. ##
## .. .. .. .. .. ## .. .. .. .. ## .. .. ##
## .. .. .. .. .. ## .. ## ## ## ## ## .. ##
## .. .. .. .. .. .. .. .. .. .. .. .. .. ##
## .. .. .. .. .. .. .. .. .. .. .. .. .. ##
## ## ## ## ## ## ## ## ## ## ## ## ## ## ##
`;

  const items = `
Ta 2,2
Tb 3,2
`;

  const actors = `
@1 7,8
@2 13,8
@3 12,8
`;

  const world = new World({
    w: 15,
    h: 10,
    wcell: 32,
    tsItems: tsIndoor,
    tsActors: tsChars,
    tsTerrainLayers: [tsRpg, tsRpg, tsIndoor],
  });

  world.SetTerrain([terrainLayer0, terrainLayer1, terrainLayer2]);
  world.SetZones(zones);
  world.SetItems(items);
  world.SetActors(actors);
  world.SetEgo("@3", {x:8, y:3});

  world.Render();
  world.obuf.CopyTo(ctx);

  world.addEventListener("changed", function(e) {
    world.obuf.CopyTo(ctx);
  });

  window.addEventListener("keydown", function(e:KeyboardEvent) {
    //if (e.ctrlKey == true && e.key == "c") {
    //  Log("break");
    //  return;
    //}

    switch(e.key) {
    case "ArrowLeft":
      world.ProcessCmd("ego h");
      e.preventDefault();
      break;
    case "ArrowRight":
      world.ProcessCmd("ego l");
      e.preventDefault();
      break;
    case "ArrowUp":
      world.ProcessCmd("ego k");
      e.preventDefault();
      break;
    case "ArrowDown":
      world.ProcessCmd("ego j");
      e.preventDefault();
      break;
    }
  });
}


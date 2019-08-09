import * as data from "./data";

interface Room {
  id: string,
  name: string,
  width: number,
  height: number,
  wcell: number,                // cell size in pixels
  terrainTilesheets: string[],  // terrain tilesheet image urls
  itemsTilesheet: string,       // items tilesheet image url
  actorsTilesheet: string,      // characters tilesheet image url
  terrainlayers: string[][],
  zones: string[],
  items: string[],
  actors: string[],
}

type RoomsMap={[id:string]:Room};

interface RoomInfoReq {
  id: string,

  // List of requested items
  // Available for request:
  //   "desc"
  //   "settings"
  //   "terrainlayers"
  //   "zones"
  //   "items"
  //   "actors"
  what: string[],   
}

interface RoomInfo {
  id: string,
  name?: string,
  width?: number,
  height?: number,
  wcell?: number,                // cell size in pixels
  terrainTilesheets?: string[],  // terrain tilesheet image urls
  itemsTilesheet?: string,       // items tilesheet image url
  actorsTilesheet?: string,      // characters tilesheet image url
  terrainlayers?: string[][],
  zones?: string[],
  items?: string[],
  actors?: string[],
}

function IsRoomInfoReq(req:RoomInfoReq):boolean {
  if (typeof req.id === "string" && Array.isArray(req.what)) {
    return true;
  }
  return false;
}

function LoadRoomsData(dirname:string):RoomsMap {
  let rooms:RoomsMap = {};

  data.ProcessJsonFiles(dirname, /^room.*\.json$/, function(room:Room) {
    // The following properties should be defined in .json file:
    //   id
    //   width, height
    //   wcell
    //   terrainTilesheets
    //   itemsTilesheet
    //   actorsTilesheet
    //   terrainlayers
    //   zones
    if (typeof room.id !== "string" ||
        room.id.length == 0 ||
        typeof room.width !== "number" ||
        typeof room.height !== "number" ||
        typeof room.wcell !== "number" ||
        !Array.isArray(room.terrainTilesheets) ||
        typeof room.itemsTilesheet !== "string" ||
        typeof room.actorsTilesheet !== "string" ||
        !Array.isArray(room.terrainlayers) ||
        !Array.isArray(room.zones)) {
      return;
    }

    room.items = room.items || [];
    room.actors = room.actors || [];

    rooms[room.id] = room;
  });

  return rooms;
}

function contains(items:any[], item:any) {
  return items.indexOf(item) != -1;
}

function CreateRoomInfo(room:Room, what:string[]):RoomInfo {
  let rminfo = <RoomInfo>{};
  rminfo.id = room.id;

  if (contains(what, "desc")) {
    rminfo.name = room.name;
    rminfo.width = room.width;
    rminfo.height = room.height;
    rminfo.wcell = room.wcell;
  }
  if (contains(what, "settings")) {
    rminfo.name = room.name;
    rminfo.width = room.width;
    rminfo.height = room.height;
    rminfo.wcell = room.wcell;
    rminfo.terrainTilesheets = room.terrainTilesheets;
    rminfo.itemsTilesheet = room.itemsTilesheet;
    rminfo.actorsTilesheet = room.actorsTilesheet;
  }
  if (contains(what, "terrainlayers")) {
    rminfo.terrainlayers = room.terrainlayers;
  }
  if (contains(what, "zones")) {
    rminfo.zones = room.zones;
  }
  if (contains(what, "items")) {
    rminfo.items = room.items;
  }
  if (contains(what, "actors")) {
    rminfo.actors = room.actors;
  }

  return rminfo;
}

function AllRoomIDs(rooms:RoomsMap):string[] {
  let ids:string[] = [];
  for (const id of Object.keys(rooms)) {
    ids.push(id);
  }
  return ids;
}

export {
  Room,
  RoomsMap,
  RoomInfoReq,
  RoomInfo,
  IsRoomInfoReq,
  LoadRoomsData,
  CreateRoomInfo,
  AllRoomIDs,
};


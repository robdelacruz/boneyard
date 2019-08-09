// Room data structures
// Need to keep Room structures in sync with server/room.ts
// There must be a better way to do this.
//

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

function BlankRoom():Room {
  let room = <Room>{};
  room.id = "";
  room.name = "";
  room.width = 0;
  room.height = 0;
  room.wcell = 0;
  room.terrainTilesheets = [];
  room.itemsTilesheet = "";
  room.actorsTilesheet = "";
  room.terrainlayers = [];
  room.zones = [];
  room.items = [];
  room.actors = [];

  return room;
}

function SetRoomInfo(room:Room, rminfo:RoomInfo) {
  if (rminfo.name != null) {
    room.name = rminfo.name;
  }
  if (rminfo.width != null) {
    room.width = rminfo.width;
  }
  if (rminfo.height != null) {
    room.height = rminfo.height;
  }
  if (rminfo.wcell != null) {
    room.wcell = rminfo.wcell;
  }
  if (rminfo.terrainTilesheets != null) {
    room.terrainTilesheets = rminfo.terrainTilesheets;
  }
  if (rminfo.itemsTilesheet != null) {
    room.itemsTilesheet = rminfo.itemsTilesheet;
  }
  if (rminfo.actorsTilesheet != null) {
    room.actorsTilesheet = rminfo.actorsTilesheet;
  }
  if (rminfo.terrainlayers != null) {
    room.terrainlayers = rminfo.terrainlayers;
  }
  if (rminfo.zones != null) {
    room.zones = rminfo.zones;
  }
  if (rminfo.items != null) {
    room.items = rminfo.items;
  }
  if (rminfo.actors != null) {
    room.actors = rminfo.actors;
  }
}

export {
  Room,
  RoomsMap,
  RoomInfoReq,
  RoomInfo,
  BlankRoom,
  SetRoomInfo,
};


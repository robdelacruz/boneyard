import {Pos,Rect,Log} from "./util";
import {Sprite} from "./sprite";
import {TileSheet,TileIndex} from "./tile";
import {OffBuf,NewOffBuf} from "./offbuf";
import {Grid} from "./grid";

interface WorldDefn {
  w: number,
  h: number,
  wcell: number,
  tsItems: TileSheet,
  tsActors: TileSheet,
  tsTerrainLayers: TileSheet[],
}

type FnEventHandler={(e:any):void};

class World {
  // Number of cells horizontally and vertically
  w: number;
  h: number;

  // Length of one cell in pixels
  wcell: number;

  // TileSheets for terrain layers
  tsTerrainLayers: TileSheet[];

  // TileSheets for items, characters
  tsItems: TileSheet;
  tsActors: TileSheet;

  // Define walkable areas.
  // Unlike terrain layers, zones are represented internally
  // as rows of zone cells, each zone cell represented as a string.
  zoneRows: string[][];

  // Terrain layers tiles.
  terrainLayers:string[][];

  // Mobile items.
  items: Sprite[];

  // Your character.
  ego: Sprite;

  // Other characters.
  actors: Sprite[];

  // Grid and offscreen buffer
  grid: Grid;
  obuf: OffBuf;

  // Event listeners
  // map of callback functions
  eventListeners: {[id:string]:FnEventHandler[]};

  constructor(defn:WorldDefn) {
    this.w = defn.w;
    this.h = defn.h;
    this.wcell = defn.wcell;

    this.tsItems = defn.tsItems;
    this.tsActors = defn.tsActors;
    this.tsTerrainLayers = defn.tsTerrainLayers;

    this.zoneRows = [];
    this.items = [];
    this.actors = [];
    this.terrainLayers = [];
    this.eventListeners = {};

    // Temp actor until SetEgo() sets the right one.
    this.ego = new Sprite(this.tsActors, "");

    // Grid and Offscreen buffer for world rendering
    this.obuf = NewOffBuf(this.w*this.wcell, this.h*this.wcell);
    this.grid = new Grid(this.obuf.cv, this.wcell);
  }

  addEventListener(eventID:string, cb:FnEventHandler) {
    if (this.eventListeners[eventID] == null) {
      this.eventListeners[eventID] = [cb];
      return;
    }

    this.eventListeners[eventID].push(cb);
  }

  fireEvent(eventID:string, e:any=null) {
    if (this.eventListeners[eventID] == null) {
      return;
    }

    for (const cb of this.eventListeners[eventID]) {
      cb(e);
    }
  }

  SetZones(zonesBlk:string) {
    let zoneRows = [];

    const srows = parseBlockToRows(zonesBlk);
    for (const srow of srows) {
      const rowCells = srow.split(/\s+/);
      zoneRows.push(rowCells);
    }

    this.zoneRows = zoneRows;
  }

  SetTerrain(terrainLayerBlks:string[]) {
    for (const layerBlk of terrainLayerBlks) {
      this.terrainLayers.push(parseBlockToRows(layerBlk));
    }
  }

  SetItems(itemsBlk:string) {
    this.items = parseSpriteBlock(itemsBlk, this.tsItems);
  }

  SetActors(actorsBlk:string) {
    this.actors = parseSpriteBlock(actorsBlk, this.tsActors);
  }

  SetEgo(ti:TileIndex, pos:Pos) {
    this.ego = new Sprite(this.tsActors, ti);
    this.ego.x = pos.x;
    this.ego.y = pos.y;
  }

  Render() {
    renderTerrain(this.grid, this.tsTerrainLayers, this.terrainLayers);
    renderItems(this.grid, this.tsItems, this.items); 
    renderItems(this.grid, this.tsActors, this.actors); 
    renderItems(this.grid, this.tsActors, [this.ego]);
  }

  ProcessCmd(cmd:string) {
    // Available commands:
    //   ego [hjkl]

    const match = cmd.match(/^ego\s+([hjkl])/);
    if (match == null || match.length <= 1) {
      return;
    }

    const dir = match[1];
    if (this.MoveEgo(dir)) {
      this.Render();
      this.fireEvent("changed");
    }
  }

  zoneCell(x:number, y:number):string {
    if (this.zoneRows.length-1 < y) {
      return "##";
    }

    const zoneCells = this.zoneRows[y];
    if (zoneCells.length-1 < x) {
      return "##";
    }
    return zoneCells[x];
  }

  // Return whether actor can walk over the cell.
  // Walkable cell should have zone cell starting with "."
  // and no items or actors occupying it.
  walkCell(x:number, y:number):boolean {
    const zoneCell = this.zoneCell(x,y);
    if (zoneCell[0] != ".") {
      return false;
    }

    for (const spItem of this.items) {
      if (spItem.x == x && spItem.y == y) {
        return false;
      }
    }
    for (const spActor of this.actors) {
      if (spActor.x == x && spActor.y == y) {
        return false;
      }
    }
    return true;
  }

  MoveEgo(dir:string):boolean {
    let targetx = this.ego.x;
    let targety = this.ego.y;

    switch (dir) {
    case "h":
      targetx--;
      break;
    case "l":
      targetx++;
      break;
    case "j":
      targety++;
      break;
    case "k":
      targety--;
      break;
    }

    // No movement
    if (this.ego.x == targetx && this.ego.y == targety) {
      return false;
    }

    // Can walk to target cell
    if (this.walkCell(targetx, targety)) {
      this.ego.x = targetx;
      this.ego.y = targety;
      return true;
    }
    return false;
  }
}

// Given a block of strings,
// filter out empty rows and extra whitespace at the ends of each row.
function parseBlockToRows(block:string):string[] {
  let srows = block.split("\n");
  srows = srows.filter(srow => srow.trim().length > 0);

  let retRows = [];
  for (const srow of srows) {
    retRows.push(srow.trim());
  }
  return retRows;
}

function parseSpriteBlock(itemsBlk:string, ts:TileSheet):Sprite[] {
  // Each sprite item is defined in a separate line:
  // <tile index alias> <xloc,yloc>
  //
  // Ex.
  // chair1 20,5
  // table1 21,5
  // chair2 22,5
  //

  let spItems = [];

  const itemLines = parseBlockToRows(itemsBlk);
  for (const sline of itemLines) {
    // <alias> <xloc,yloc>   ignore any extra whitespace
    let toks = sline.split(" ");
    toks = toks.filter(tok => tok.trim().length > 0);

    // Ignore any invalid item lines.
    if (toks.length <= 1) {
      Log(`Skipping invalid item line: ${sline}`);
      continue;
    }

    let alias = toks[0];
    let loc = toks[1];
    let locToks = loc.split(",");
    if (locToks.length <= 1) {
      Log(`Skipping invalid item line: ${sline}`);
      continue;
    }

    let sxloc = locToks[0];
    let syloc = locToks[1];
    let xloc = parseInt(sxloc);
    let yloc = parseInt(syloc);
    if (xloc === NaN || yloc === NaN) {
      Log(`Skipping invalid item line: ${sline}`);
      continue;
    }

    // item: alias, xloc, yloc
    const spItem = new Sprite(ts, alias);
    spItem.x = xloc;
    spItem.y = yloc;
    spItems.push(spItem);
  }

  return spItems;
}

function renderTerrain(grid:Grid, tss:TileSheet[], terrainLayers:string[][]) {
  for (let i=0; i < terrainLayers.length; i++) {
    renderTerrainLayer(grid, tss[i], terrainLayers[i]);
  }
}

// Render one layer of terrain tiles to grid.
// terrainRows represents the tiles, each tile whitespace delimited.
function renderTerrainLayer(grid:Grid, ts:TileSheet, terrainRows:string[]) {
  for (let y=0; y < terrainRows.length; y++) {
    const toks = terrainRows[y].split(/\s+/);
    for (let x=0; x < toks.length; x++) {
      const tiAlias = toks[x];
      if (tiAlias[0] == "-") {
        continue;
      }
      grid.DrawTile(ts, tiAlias, x,y);
    }
  }
}

function renderItems(grid:Grid, ts:TileSheet, items:Sprite[]) {
  for (const spItem of items) {
    grid.DrawSprite(spItem, spItem.x, spItem.y);
  }
}

export {
  World,
};

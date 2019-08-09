class Kb {
  // Map indicating whether a key is 'active' (currently held down)
  activeKeys: {[key:string]:boolean};

  constructor() {
    this.activeKeys = {};
  }

  HandleKBEvent(e:KeyboardEvent) {
    // keydown sets the key as active,
    // keyup removes the key from the map.
    if (e.type == "keydown") {
      this.activeKeys[e.key] = true;
    } else if (e.type == "keyup") {
      delete this.activeKeys[e.key];
    }
  }

  RemoveKey(key:string) {
    delete this.activeKeys[key];
  }

  // Returns all active keys.
  ActiveKeys():string[] {
    return Object.keys(this.activeKeys);
  }

  // Return whether key is active (held down)
  IsKeyDown(key:string):boolean {
    if (this.ActiveKeys().indexOf(key) != -1) {
      return true;
    }
    return false;
  }
}

export {
  Kb,
};




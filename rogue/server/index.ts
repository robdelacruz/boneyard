import path from "path";
import express from "express";
import http from "http";
import socketio from "socket.io";
import {Log} from "./util";
import * as data from "./data";
import * as rmlib from "./room";

// ***
// _rooms: All available rooms
// _rooms[<room id>] = <Room>
// ***
let _rooms:rmlib.RoomsMap = {};

function main() {
  _rooms = rmlib.LoadRoomsData("dist/server/data");

  const app = express();
  const server = new http.Server(app);
  const io = socketio(server);

  app.use("/rogue", express.static("public"));
  app.get("/room", GET_Rooms);
  app.get("/room/:id", GET_Room);

  io.on("connection", newSockIOConnection);

  server.listen(3000, function() {
    Log("Listening...");
  });
}

main();

function GET_Rooms(req:express.Request, res:express.Response) {
  let roomInfos = [];
  for (const id of Object.keys(_rooms)) {
    const room = _rooms[id];
    const roomInfo = rmlib.CreateRoomInfo(room, ["desc"]);
    roomInfos.push(roomInfo);

  }
  res.json(roomInfos);
}

function GET_Room(req:express.Request, res:express.Response) {
  const id = req.params.id;

  const room = _rooms[id];
  if (room == null) {
    res.status(404).send("Room not found");
    return;
  }

  const roomInfo = rmlib.CreateRoomInfo(room,
    ["settings", "terrainlayers", "zones", "items", "actors"]);
  res.json(roomInfo);
}

function newSockIOConnection(sock:socketio.Socket) {
  Log("A client connected.");

  sock.on("disconnect", onSockIODisconnect);

  // Client sends 'Enter room <id>'.
  // sock.on("enter room ns", function(req:<enter room struct>) {}
  //   sock.emit("ns", <room info struct>)
}

function onSockIODisconnect() {
  Log("Client disconnected.");
}



import fs from "fs";
import {Log} from "./util";

// Read and process all json text files from given directory.
function ProcessJsonFiles(dirname:string, matchexp:RegExp, mapfn:(o:any)=>void) {
  let files:string[] = [];
  try {
    files = fs.readdirSync(dirname);
  } catch(e) {
    Log(e);
    return;
  }

  const matchFiles = files.filter(file => file.match(matchexp) != null);
  for (const file of matchFiles) {
    try {
      const contents = fs.readFileSync(`${dirname}/${file}`, "utf8");
      const o = JSON.parse(contents);
      mapfn(o);
    } catch(e) {
      // Skip file if any error while opening or parsing.
      continue;
    }
  }
}

export {
  ProcessJsonFiles,
};


export {
  Pos,
  Rect,
  PosT,
  RectT,
  Log,
  LoadImage,
  LoadImages,
  LoadImageCB,
  LoadImagesCB,
  ImgMap
};

interface Pos {
  x: number;
  y: number;
}

interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
}

type PosT=[number, number];
type RectT=[number, number, number, number];

function Log(s: string) {
  console.log(s);
}
//
// Async load image file, callback return with img element.
interface LoadImageCB { (img: HTMLImageElement): void }
function LoadImage(imageFile: string, cb: LoadImageCB) {
  let img = document.createElement("img");
  img.addEventListener("load", function() {
    cb(img);
  });
  img.src = imageFile;
}

// Async load image files, callback return with map of img elements.
interface ImgMap { [imageFile: string]: HTMLImageElement }
interface LoadImagesCB { (imgMap: ImgMap): void }
function LoadImages(imageFiles: string[], cb: LoadImagesCB) {
  let imgMap: ImgMap = {};
  let nLoaded = 0;

  for (const imageFile of imageFiles) {
    let img = document.createElement("img");
    img.src = imageFile;
    img.addEventListener("load", function() {
      nLoaded++;
      if (nLoaded == imageFiles.length) {
        cb(imgMap);
      }
    });
    imgMap[imageFile] = img;
  }
}


// Canvas utility functions

class OffBuf {
  cv: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;
  w: number;
  h: number;

  constructor(cv:HTMLCanvasElement) {
    this.cv = cv;
    this.ctx = cv.getContext("2d")!;
    this.w = cv.width;
    this.h = cv.height;
  }

  CopyTo(ctxDest:CanvasRenderingContext2D) {
    ctxDest.drawImage(this.cv, 0,0);
  }
}

function NewOffBufCanvas(cvBase:HTMLCanvasElement):OffBuf {
  // Create offscreen canvas same dimensions as cvBase.
  const cv = document.createElement("canvas");
  cv.width = cvBase.width;
  cv.height = cvBase.height;

  return new OffBuf(cv);
}

function NewOffBuf(w:number, h:number):OffBuf {
  // Create offscreen canvas with given dimensions.
  const cv = document.createElement("canvas");
  cv.width = w;
  cv.height = h;

  return new OffBuf(cv);
}

export {
  OffBuf,
  NewOffBufCanvas,
  NewOffBuf,
};


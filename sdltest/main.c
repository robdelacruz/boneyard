// Globals
// -------
// Cv (Canvas*)
//
// Types
// -----
// EventType
// Graphics
//
// Functions
// ---------
// init()
// render()
// update()
// cleanup()
// handleEvent() Event
//

#include <stdio.h>
#include <stdlib.h>
#include "util.h"
#include "graphics.h"
#include "canvas.h"
#include "sprite.h"
#include "event.h"

int init();
void render();
void update();
void cleanup();
Event handleEvent();

Canvas *Cv;
Canvas *CvImg;
Canvas *CvRGB;
SpriteSheet *Ss1;
Sprite *Spr1;

int main(int argc, char *argv[]) {
    int errn = 0;

    errn = init();
    if (errn != 0) {
        cleanup();
        return 1;
    }

    unsigned int lastFrameTick = 0;
    unsigned int curTick = 0;
    //int msPerFrame = 1000 / 60; // 60 fps
    int msPerFrame = 1000 / 10;
    //int msPerFrame = 0;

    while (1) {
        Event e = handleEvent();
        if (e.type == ET_QUIT) {
            break;
        }

        curTick = SDL_GetTicks();
        if (curTick > lastFrameTick + msPerFrame) {
            lastFrameTick = curTick;

            update();
            render();
        }
    }

    cleanup();

    return 0;
}

int init() {
    int errn = InitGraphics();
    if (errn < 0) {
        return 1;
    }

    //char *imgfile = "assets/lky.bmp";
    char *imgfile = "assets/inv_transparent.png";
    CvImg = Canvas_CreateFromImg(imgfile);
    if (CvImg == NULL) {
        return 1;
    }

    Rect bmpSize = Canvas_Size(CvImg);
    Cv = Canvas_CreateWin("robwin", bmpSize.W, bmpSize.H);
    if (Cv == NULL) {
        return 1;
    }

    CvRGB = Canvas_CreateRGB(640, 480);
    if (CvRGB == NULL) {
        return 1;
    }

    //Canvas_Clear(Cv, 0x000000FF);
    Canvas_Clear(Cv, 0x222200FF);
    Canvas_Clear(CvRGB, 0x000000FF);

    Ss1 = SpriteSheet_CreateFromImg("assets/scottpilgrim.png", NullRect(), NewRect(0,0,108,140));
    int wx = 38;
    //Ss1 = SpriteSheet_CreateFromImg("assets/tintin2.png", NewRect(0, 70, wx*11, 70*2), NewRect(0,0,wx,70));
    Spr1 = Sprite_Create(Ss1);
    SpriteAnimation sa = NewSpriteAnimation("idle", 0, 8);
    Sprite_AddAnimation(Spr1, sa);
    Sprite_Animate(Spr1, "idle");

    return 0;
}

void cleanup() {
    Canvas_Destroy(Cv);
    Canvas_Destroy(CvImg);
    SpriteSheet_Destroy(Ss1);
    Sprite_Destroy(Spr1);

    Cv = NULL;
    CvImg = NULL;
    Ss1 = NULL;
    Spr1 = NULL;

    QuitGraphics();
}


void update() {
    Sprite_NextFrame(Spr1);
}

void renderImg() {
    int x = rand() % 640;
    int y = rand() % 480;

    int r = rand() % 0xFF;
    int g = rand() % 0xFF;
    int b = rand() % 0xFF;
    Color fg = NewColor(r, g, b, 0xFF);

    Canvas_PutPixel(Cv, x, y, ColorWFrom(fg));

    int maxW = 20;
    int maxH = 20;
    Rect srcSize = Canvas_Size(CvImg);
    Rect destSize = Canvas_Size(Cv);

    Rect srcRect, destRect;
    srcRect.X = rand() % (srcSize.W - maxW);
    srcRect.Y = rand() % (srcSize.H - maxH);
    destRect.X = rand() % (destSize.W - maxW);
    destRect.Y = rand() % (destSize.H - maxH);

    int w = rand() % maxW;
    int h = rand() % maxH;
    srcRect.W = destRect.W = w;
    srcRect.H = destRect.H = h;
    Canvas_BitBlt(Cv, srcRect, CvImg, srcRect);

    Canvas_RenderUpdate(Cv);
}

void renderSS() {
    Canvas *cvSS = Ss1->Cv;
    Rect cvSize = Canvas_Size(Cv);
    Rect ssSize = Canvas_Size(cvSS);

    int x = rand() % cvSize.W - ssSize.W;
    int y = rand() % cvSize.H - ssSize.H;
    Rect destRect = NewRect(x, y, ssSize.W, ssSize.H);
    Canvas_BitBlt(Cv, destRect, cvSS, ssSize);

    Canvas_RenderUpdate(Cv);
}

static int xSpr = 0;
static int ySpr = 0;

void renderSpr() {
    xSpr += 10;

    Rect cvSize = Canvas_Size(Cv);
    if (xSpr < 0) {
        xSpr = 0;
    }
    if (ySpr < 0) {
        ySpr = 0;
    }
    if (xSpr > cvSize.W-1) {
        xSpr = 0;
        ySpr += 10;
    }
    if (ySpr > cvSize.H-1) {
        xSpr = 0;
        ySpr = 0;
    }

    Sprite_Draw(Spr1, Cv, xSpr, ySpr);
}

void render() {
//    renderImg();
//    renderSS();
    Canvas_Clear(Cv, 0x222200FF);
    renderSpr();
    Canvas_RenderUpdate(Cv);
}

Event handleEvent() {
    Event e = PollEvent();
    return e;
}



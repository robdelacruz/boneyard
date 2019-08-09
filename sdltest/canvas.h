#ifndef __CANVAS__
#define __CANVAS__

#include <SDL.h>
#include <SDL_image.h>
#include "graphics.h"

typedef struct {
    SDL_Surface *Sur;
    SDL_Renderer *Ren;
    SDL_Window *Win;
} Canvas;

Canvas *Canvas_Init();
Canvas *Canvas_CreateWin(char *title, int w, int h);
Canvas *Canvas_CreateFromImg(char *file);
Canvas *Canvas_CreateRGB(int w, int h);
void Canvas_Destroy(Canvas *cv);
Rect Canvas_Size(Canvas *cv);
void Canvas_Clear(Canvas *cv, ColorW clrW);
void Canvas_RenderUpdate(Canvas *cv);
void Canvas_PutPixel(Canvas *cv, int x, int y, ColorW clrW);
void Canvas_BitBlt(Canvas *cv, Rect destRect, Canvas *cvSrc, Rect srcRect);

#endif


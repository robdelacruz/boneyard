// Types
// -----
// Color
// ColorW
//
// Functions
// ---------
// InitGraphics() int
// QuitGraphics()
// NewColor(r, g, b, a) Color
// ColorFromW(ColorW clrW) Color
// ColorWFrom(Color clr) ColorW
// NewRect(int x, int y, int w, int h) Rect
// NullRect() Rect
// IsNullRect(Rect rect) int
//

#include <SDL.h>
#include "graphics.h"

int InitGraphics() {
    int errn = SDL_Init(SDL_INIT_EVERYTHING);
    return errn;
}

int QuitGraphics() {
    SDL_Quit();
}

Color NewColor(int r, int g, int b, int a) {
    Color clr = {r, g, b, a};
    return clr;
}
Color ColorFromW(ColorW clrW) {
    int r = (clrW & 0xFF000000) >> 24;
    int g = (clrW & 0x00FF0000) >> 16;
    int b = (clrW & 0x0000FF00) >> 8;
    int a = (clrW & 0x000000FF) >> 0;

    return NewColor(r, g, b, a);
}
ColorW ColorWFrom(Color clr) {
    return ((clr.R & 0x000000FF) << 24)
        | ((clr.G & 0x000000FF) << 16)
        | ((clr.B & 0x000000FF) << 8)
        | ((clr.A & 0x000000FF) << 0);
}

Rect NewRect(int x, int y, int w, int h) {
    Rect rect = {x, y, w, h};
    return rect;
}

Rect NullRect() {
    return NewRect(-1, -1, -1, -1);
}

int IsNullRect(Rect rect) {
    return rect.X < 0;
}




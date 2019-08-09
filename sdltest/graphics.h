#ifndef __GRAPHICS__
#define __GRAPHICS__

#include <stdint.h>

typedef struct {
    int R;  // red
    int G;  // green
    int B;  // blue
    int A;  // alpha
} Color;

typedef int32_t ColorW;

typedef struct {
    int X;
    int Y;
    int W;
    int H;
} Rect;

int InitGraphics();
int QuitGraphics();
Color NewColor(int r, int g, int b, int a);
Color ColorFromW(ColorW clrW);
ColorW ColorWFrom(Color clr);
Rect NewRect(int x, int y, int w, int h);
Rect NullRect();
int IsNullRect(Rect rect);

#endif


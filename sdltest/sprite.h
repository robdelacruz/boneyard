#ifndef __SPRITE__
#define __SPRITE__

// Types
// -----
// Sprite
// SpriteSheet
// SpriteAnimation
// SpriteCell

#include "canvas.h"

// SpriteSheet defines a rectangular image containing a fixed number of 
// cells of the same dimension. The cells are sequenced together in 
// a number of animation sequences.
typedef struct {
    Canvas *Cv;
    Rect ContentRect;
    Rect CellSize;
    int NCells;
} SpriteSheet;

// SpriteCell is a width x height array of pixels.
typedef struct {
    ColorW *Pixels;
    int W;
    int H;
} SpriteCell;

// SpriteAnimation defines an animation sequence - each frame is a cell
// from the spritesheet.
typedef struct {
    char *Name;
    int StartIdx;
    int NCells;
} SpriteAnimation;

// For now, support a fixed maximum number of animations per sprite.
// In the future, we can make the number of animtions dynamic.
#define SPRITE_MAX_ANIMATIONS 10

typedef struct {
    SpriteSheet *SS;
    SpriteAnimation Animations[SPRITE_MAX_ANIMATIONS]; 
    SpriteAnimation *curAnimation;
    int iCurCell;
} Sprite;

SpriteSheet* SpriteSheet_Create(Canvas *cv, Rect contentRect, Rect cellSize);
SpriteSheet* SpriteSheet_CreateFromImg(char *file, Rect contentRect, Rect cellSize);
void SpriteSheet_Destroy(SpriteSheet *ss);

Sprite* Sprite_Init();
Sprite* Sprite_Create(SpriteSheet *ss);
void Sprite_Destroy(Sprite *spr);
SpriteAnimation NewSpriteAnimation(char *name, int startIdx, int nCells);
int Sprite_AddAnimation(Sprite *spr, SpriteAnimation saAdd);
void Sprite_Draw(Sprite *spr, Canvas *destCv, int x, int y);
void Sprite_NextFrame(Sprite *spr);
void Sprite_ResetFrame(Sprite *spr);
void Sprite_Animate(Sprite *spr, char *animationName);

#endif


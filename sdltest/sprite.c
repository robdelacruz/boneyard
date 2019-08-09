// Types
// -----
// Sprite
// SpriteSheet
// SpriteAnimation
// SpriteCell
//
// Local functions
// ---------------
//
// Functions
// ---------
// SpriteSheet_Init() SpriteSheet*
// SpriteSheet_Create(Canvas *cv, Rect contentRect, Rect cellSize)
// SpriteSheet_CreateFromCells(Rect cellSize, SpriteCell cells[], int nCells)
// SpriteSheet_CreateFromImg(char *file, Rect contentRect, Rect cellSize)
// SpriteSheet_Destroy(SpriteSheet *ss)
//
// Sprite_Init() Sprite*
// Sprite_Create(SpriteSheet *ss) Sprite*
// Sprite_Destroy(Sprite *spr)
// NewSpriteAnimation(char *name, int startIdx, int nCells) SpriteAnimation
// Sprite_AddAnimation(Sprite *spr, SpriteAnimation saAdd) int
//
// Sprite_Draw(Sprite *spr, Canvas *destCv, int x, int y)
// Sprite_NextFrame(Sprite *spr)
// Sprite_ResetFrame(Sprite *spr)
// Sprite_Animate(Sprite *spr, char *animationName)
//

#include <stdio.h>
#include <string.h>
#include "util.h"
#include "canvas.h"
#include "sprite.h"

//
// SpriteSheet
//
SpriteSheet* SpriteSheet_Init() {
    SpriteSheet *ss = Malloc(sizeof(SpriteSheet));
    ss->Cv = NULL;
    ss->CellSize = NewRect(0, 0, 0, 0);
    ss->NCells = 0;

    return ss;
}

int numCellsInSheet(Rect cellSize, Rect sheetSize) {
    if (sheetSize.W < cellSize.W || sheetSize.H < cellSize.H) {
        return 0;
    }

    int nCellsInRow = sheetSize.W / cellSize.W;
    int nRows = sheetSize.H / cellSize.H;
    return nCellsInRow * nRows;
}

SpriteSheet* SpriteSheet_Create(Canvas *cv, Rect contentRect, Rect cellSize) {
    SpriteSheet *ss = SpriteSheet_Init();
    ss->Cv = cv;
    ss->CellSize = cellSize;

    ss->ContentRect = contentRect;
    if (IsNullRect(contentRect)) {
        ss->ContentRect = Canvas_Size(ss->Cv);
    }

    ss->NCells = numCellsInSheet(cellSize, ss->ContentRect);

    return ss;
}

SpriteSheet* SpriteSheet_CreateFromImg(char *file, Rect contentRect, Rect cellSize) {
    Canvas *cv = Canvas_CreateFromImg(file);
    SpriteSheet *ss = SpriteSheet_Create(cv, contentRect, cellSize);

    return ss;
}

// Draw cell starting from (offsetX,offsetY) to canvas.
void drawCell(Canvas *cv, SpriteCell cell, int offsetX, int offsetY) {
    for (int cellY = 0; cellY < cell.H; cellY++) {
        for (int cellX = 0; cellX < cell.W; cellX++) {
            ColorW clrW = cell.Pixels[(cellY * cell.W) + cellX];
            Canvas_PutPixel(cv, offsetX + cellX, offsetY + cellY, clrW);
        }
    }
}

SpriteSheet* SpriteSheet_CreateFromCells(Rect cellSize, SpriteCell cells[], int nCells) {
    SpriteSheet *ss = SpriteSheet_Init();

    // Create canvas big enough to fit all the cells.
    // Cells are arranged horizontally from left to right.
    int sheetW = cellSize.W * nCells;
    int sheetH = cellSize.H;
    ss->Cv = Canvas_CreateRGB(sheetW, sheetH);
    ss->CellSize = cellSize;
    ss->NCells = nCells;

    for (int i=0; i < nCells; i++) {
        SpriteCell cell = cells[i];

        int offsetX = cell.W * i;
        int offsetY = 0;
        drawCell(ss->Cv, cell, offsetX, offsetY);
    }

    return ss;
}

void SpriteSheet_Destroy(SpriteSheet *ss) {
    Canvas_Destroy(ss->Cv);
    ss->Cv = NULL;
    Free(ss);
}

SpriteAnimation NewSpriteAnimation(char *name, int startIdx, int nCells) {
    SpriteAnimation sa = {name, startIdx, nCells};
    return sa;
}

//
// Sprite
//
Sprite* Sprite_Init() {
    Sprite *spr = Malloc(sizeof(Sprite));
    spr->SS = NULL;
    spr->curAnimation = NULL;
    spr->iCurCell = -1;

    for (int i=0; i < SPRITE_MAX_ANIMATIONS; i++) {
        spr->Animations[i] = NewSpriteAnimation(NULL, 0, 0);
    }

    return spr;
}

Sprite* Sprite_Create(SpriteSheet *ss) {
    Sprite *spr = Sprite_Init();
    spr->SS = ss;

    return spr;
}

void Sprite_Destroy(Sprite *spr) {
    spr->SS = NULL;
    spr->curAnimation = NULL;
    Free(spr);
}

int Sprite_AddAnimation(Sprite *spr, SpriteAnimation saAdd) {
    if (saAdd.Name == NULL) {
        Log("Can't add sprite animation without a name.");
        return 0;
    }

    // Validate that saAdd.StartIdx and NFrames are within frame range.
    if (saAdd.StartIdx > spr->SS->NCells-1) {
        Log("Sprite_AddAnimation(): '%s' start index %d out of range.", saAdd.Name, saAdd.StartIdx);
        return 0;
    }

    if (saAdd.StartIdx + saAdd.NCells - 1 > spr->SS->NCells-1) {
        Log("Sprite_AddAnimation(): '%s' NCells %d out of range.", saAdd.Name, saAdd.NCells);
        return 0;
    }

    for (int i=0; i < SPRITE_MAX_ANIMATIONS; i++) {
        SpriteAnimation sa = spr->Animations[i];
        if (sa.Name == NULL || strcmp(sa.Name, saAdd.Name) == 0) {
            spr->Animations[i] = saAdd;
            return 1;
        }
    }

    return 0;
}

void Sprite_Draw(Sprite *spr, Canvas *destCv, int x, int y) {
    if (spr->curAnimation == NULL || spr->iCurCell == -1) {
        Log("Sprite_Draw(): No animation selected.");
        return;
    }

    SpriteAnimation *sa = spr->curAnimation;
    int iCell = spr->iCurCell;
    if (iCell < sa->StartIdx || iCell > sa->StartIdx + sa->NCells - 1) {
        Log("Sprite_Draw(): cell %d is out of range. Fixing..", iCell);
        spr->iCurCell = sa->StartIdx;
        iCell =spr->iCurCell;
    }


    // Get current cell rect
    SpriteSheet *ss = spr->SS;
    int nCells = numCellsInSheet(ss->CellSize, ss->ContentRect);
    if (iCell > nCells-1) {
        Log("Sprite_Draw(): cell %d is out of sheet range. Skipping draw.", iCell);
        return;
    }

    int nCellsInRow = ss->ContentRect.W / ss->CellSize.W;
    int nRows = ss->ContentRect.H / ss->CellSize.H;
    int iRow = iCell / nCellsInRow;
    int iCol = iCell % nCellsInRow;
    if (iRow > nRows-1 || iCol > nCellsInRow-1) {
        Log("Sprite_Draw(): cell %d, row %d, col %d is out of range (max row: %d, max col: %d). Skipping draw.", iCell, iRow, iCol, nRows-1, nCellsInRow-1);
        return;
    }

    int xCell = ss->ContentRect.X + iCol * ss->CellSize.W;
    int yCell = ss->ContentRect.Y + iRow * ss->CellSize.H;
    Rect srcRect = NewRect(xCell, yCell, ss->CellSize.W, ss->CellSize.H);
    Rect destRect = NewRect(x, y, ss->CellSize.W-1, ss->CellSize.H);

//    Log("Sprite_Draw(): cell %d, row %d, col %d, x: %d, y: %d, w: %d, h: %d",
//        iCell, iRow, iCol, srcRect.X, srcRect.Y, srcRect.W, srcRect.H);
    Canvas_BitBlt(destCv, destRect, ss->Cv, srcRect);
}

void Sprite_NextFrame(Sprite *spr) {
    if (spr->curAnimation == NULL || spr->iCurCell == -1) {
        Log("Sprite_NextFrame(): No animation selected.");
        return;
    }

    spr->iCurCell++;
    SpriteAnimation *sa = spr->curAnimation;
    if (spr->iCurCell < sa->StartIdx ||
        spr->iCurCell > sa->StartIdx + sa->NCells - 1) {
        // Cycle back to first cell.
        spr->iCurCell = sa->StartIdx;
    }
}

void Sprite_ResetFrame(Sprite *spr) {
    if (spr->curAnimation == NULL || spr->iCurCell == -1) {
        Log("Sprite_ResetFrame(): No animation selected.");
        return;
    }

    spr->iCurCell = spr->curAnimation->StartIdx;
}

void Sprite_Animate(Sprite *spr, char *animationName) {
    for (int i=0; i < SPRITE_MAX_ANIMATIONS; i++) {
        SpriteAnimation *sa = &spr->Animations[i];
        if (sa->Name != NULL && strcmp(sa->Name, animationName) == 0) {
            spr->curAnimation = sa;
            spr->iCurCell = sa->StartIdx;
            return;
        }
    }

    Log("Sprite_Animate(): animation '%s' not found.", animationName);
}



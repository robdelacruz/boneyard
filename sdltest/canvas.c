// Types
// -----
// Canvas
//
// Local functions
// ---------------
// SDL_Rect toSDLRect(Rect rect) SDL_Rect
// setFgColor(SDL_Renderer *ren, ColorW clrW)
// createRenderer(Canvas *cv) SDL_Renderer*
//
// Functions
// ---------
// Canvas_Init() Canvas*
// Canvas_CreateWin(char *title, int w, int h) Canvas*
// Canvas_CreateFromImg(char *file) Canvas*
// Canvas_CreateRGB(int w, int h) Canvas*
// Canvas_Destroy(Canvas *cv)
// Canvas_Size(Canvas *cv) Rect
// Canvas_Clear(Canvas *cv, ColorW clrW)
// Canvas_RenderUpdate(Canvas *cv)
// Canvas_PutPixel(Canvas *cv, int x, int y, ColorW clrW)
// Canvas_BitBlt(Canvas *cv, Rect destRect, Canvas *cvSrc, Rect srcRect)
//

#include "util.h"
#include "canvas.h"

SDL_Rect toSDLRect(Rect rect) {
    SDL_Rect sdlRect = {rect.X, rect.Y, rect.W, rect.H};
    return sdlRect;
}

void setFgColor(SDL_Renderer *ren, ColorW clrW) {
    Color clr = ColorFromW(clrW);
    SDL_SetRenderDrawColor(ren, clr.R, clr.G, clr.B, clr.A);
}

SDL_Renderer *createRenderer(Canvas *cv) {
    SDL_Renderer *ren = NULL;
    const char *createfn = NULL;

    createfn = "SLD_CreateSoftwareRenderer";
    ren = SDL_CreateSoftwareRenderer(cv->Sur);
    if (ren == NULL) {
        Log("%s error (%s)", createfn, SDL_GetError());
    }

    return ren;
}

Canvas* Canvas_Init() {
    Canvas *cv = Malloc(sizeof(Canvas));
    cv->Sur = NULL;
    cv->Ren = NULL;
    cv->Win = NULL;

    return cv;
}

Canvas* Canvas_CreateWin(char *title, int w, int h) {
    Canvas *cv = Canvas_Init();

    cv->Win = SDL_CreateWindow(title, SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, w, h, SDL_WINDOW_SHOWN);
    if (cv->Win == NULL) {
        Log("SDL_CreateWindow error (%s)", SDL_GetError());
        return NULL;
    }

    cv->Sur = SDL_GetWindowSurface(cv->Win);
    if (cv->Sur == NULL) {
        Log("SDL_GetWindowSurface error (%s)", SDL_GetError());

        Canvas_Destroy(cv);
        return NULL;
    }
    cv->Ren = createRenderer(cv);

    return cv;
}

Canvas* Canvas_CreateFromImg(char *file) {
    Canvas *cv = Canvas_Init();

    cv->Sur = IMG_Load(file);
    if (cv->Sur == NULL) {
        Log("IMG_Load error (%s)", SDL_GetError());

        Canvas_Destroy(cv);
        return NULL;
    }
    cv->Ren = createRenderer(cv);

    return cv;
}

Canvas* Canvas_CreateRGB(int w, int h) {
    Canvas *cv = Canvas_Init();

    cv->Sur = SDL_CreateRGBSurface(0, w, h, 32, 0, 0, 0, 0);
    if (cv->Sur == NULL) {
        Log("SDL_CreateRGBsurface error (%s)", SDL_GetError());

        Canvas_Destroy(cv);
        return NULL;
    }
    cv->Ren = createRenderer(cv);

    return cv;
}

void Canvas_Destroy(Canvas *cv) {
    if (cv->Sur != NULL) {
        SDL_FreeSurface(cv->Sur);
    }
    if (cv->Ren != NULL) {
        SDL_DestroyRenderer(cv->Ren);
    }
    if (cv->Win != NULL) {
        SDL_DestroyWindow(cv->Win);
    }
    cv->Sur = NULL;
    cv->Ren = NULL;
    cv->Win = NULL;
    Free(cv);
}

Rect Canvas_Size(Canvas *cv) {
    return NewRect(0, 0, cv->Sur->w, cv->Sur->h);
}

void Canvas_Clear(Canvas *cv, ColorW clrW) {
    setFgColor(cv->Ren, clrW);
    SDL_RenderClear(cv->Ren);
}

void Canvas_RenderUpdate(Canvas *cv) {
    SDL_RenderPresent(cv->Ren);
}

void Canvas_PutPixel(Canvas *cv, int x, int y, ColorW clrW) {
    setFgColor(cv->Ren, clrW);
    SDL_RenderDrawPoint(cv->Ren, x, y);
}

void Canvas_BitBlt(Canvas *cv, Rect destRect, Canvas *cvSrc, Rect srcRect) {
    SDL_Rect *src = NULL;
    SDL_Rect *dest = NULL;
    SDL_Rect sdlSrcRect, sdlDestRect;

    if (!IsNullRect(srcRect)) {
        sdlSrcRect = toSDLRect(srcRect);
        src = &sdlSrcRect;
    }
    if (!IsNullRect(destRect)) {
        sdlDestRect = toSDLRect(destRect);
        dest = &sdlDestRect;
    }

    int errn = SDL_BlitSurface(cvSrc->Sur, src, cv->Sur, dest);
    if (errn != 0) {
        Log("SDL_BlitSurface error (%s)", SDL_GetError());
    }
    if (cv->Win !=NULL) {
        SDL_UpdateWindowSurface(cv->Win);
    }
}



// Types
// -----
// EventType
// Event
//
// Functions
// ---------
// PollEvent() Event
//

#include <SDL.h>
#include "event.h"

Event PollEvent() {
    Event e = {ET_NONE, 0};

    SDL_Event ev;
    if (SDL_PollEvent(&ev)) {
        switch (ev.type) {
            case SDL_QUIT:
                e.type = ET_QUIT;
                break;
            default:
                e.type = ET_NONE;
                break;
        }
    }

    return e;
}


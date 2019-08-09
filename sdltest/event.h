#ifndef __EVENT__
#define __EVENT__

typedef enum {ET_NONE, ET_QUIT} EventType;

typedef struct {
    EventType type;
    int ch;
} Event;

Event PollEvent();

#endif


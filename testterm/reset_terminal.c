#include <unistd.h>
#include <termios.h>

void set_normal_mode() {
    struct termios term;
    tcgetattr(0, &term);
    term.c_lflag |= ECHO;
    term.c_lflag |= ICANON;
    tcsetattr(0, TCSAFLUSH, &term);
}

int main() {
    set_normal_mode();
}


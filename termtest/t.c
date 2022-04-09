#include <stdio.h>

int main(int argc, char *argv[]) {
    printf("\033[2J");
//    printf("\033[s");

    printf("\033[2;5H");
    printf("\033[0;35mhello");
    printf("\033[38;5;83mworld\n");
    printf("\033[0m");

//    printf("\033[u");

    return 0;
}


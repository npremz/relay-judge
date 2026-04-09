#include <stdio.h>
#include <stdlib.h>

void sort_the_stack(int *stack1, int len_stack1, int *stack2, int len_stack2) {
    int total = len_stack1 + len_stack2;
    int *merged = malloc((size_t) total * sizeof(int));
    int index = 0;

    if (merged == NULL) {
        return;
    }

    for (int i = 0; i < len_stack1; i++) {
        merged[index++] = stack1[i];
    }
    for (int i = 0; i < len_stack2; i++) {
        merged[index++] = stack2[i];
    }

    for (int i = 1; i < total; i++) {
        int value = merged[i];
        int j = i - 1;

        while (j >= 0 && merged[j] > value) {
            merged[j + 1] = merged[j];
            j--;
        }
        merged[j + 1] = value;
    }

    for (int i = 0; i < total; i++) {
        if (i > 0) {
            printf(" ");
        }
        printf("%d", merged[i]);
    }

    free(merged);
}

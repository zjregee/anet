#include <stdint.h>

int64_t sum(int64_t n) {
    int64_t result = 0;
    for (int64_t i = 0; i < n; i++) {
        result += i;
    }
    return result;
}

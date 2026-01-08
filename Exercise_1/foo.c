/*global shared int i = 0

main:
    spawn thread_1
    spawn thread_2
    join all threads (or wait for them to finish)
    print i

thread_1:
    do 1_000_000 times:
        i++
thread_2:
    do 1_000_000 times:
        i--
*/


// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;
pthread_mutex_t lock;


// Note the return type: void*
void* incrementingThreadFunction(){
    // TODO: increment i 1_000_000 times
    for (int j = 0; j < 1000000; j++) {
        pthread_mutex_lock (&lock);
        i++;
        pthread_mutex_unlock(&lock);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    // TODO: decrement i 1_000_000 times
    for (int k = 0; k < 1000000; k++){
        pthread_mutex_lock (&lock);
        i--;
        pthread_mutex_unlock(&lock);
    }

    return NULL;
}


int main(){
    // TODO: 
    pthread_mutex_init(&lock, NULL);
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?
    pthread_t t1, t2; 
    pthread_create(&t1, NULL, incrementingThreadFunction, NULL);
    pthread_create(&t2, NULL, decrementingThreadFunction, NULL);
    
    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    
    pthread_join(t1, NULL);
    pthread_join(t2, NULL);
    
    printf("The magic number is: %d\n", i);
    return 0;
}

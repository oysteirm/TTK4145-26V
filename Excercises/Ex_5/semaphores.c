// Initial values:
// M     = 1
// PS[2] = [0, 0]
// busy  = false

// priority: 1=high, 0=low
void allocate(int priority){
    Wait(M);
    if(busy){
        Signal(M);
        Wait(PS[priority]);
    }
    busy = true;
    Signal(M);
}

void deallocate(){
    Wait(M);
    busy = false;
    if(GetValue(PS[1]) < 0){
        Signal(PS[1]);
    } else if(GetValue(PS[0]) < 0){
        Signal(PS[0]);
    } else {
        Signal(M);
    }
}
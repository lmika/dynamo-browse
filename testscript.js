console.log("Hello world!!");

class MyTestClass {
    constructor(x, y) {
        this.x = x;
        this.y = y;
    }

    doTheSum() {
        return this.x + this.y;
    }
}


let myClass = new MyTestClass(2, 2);
console.log("My test class = " + myClass.doTheSum());


audax.ui.prompt("Hello");

class Lock {
    constructor() {
        this.locked = false;
        this.waitingQueue = [];
    }

    async acquire() {
        if (this.locked) {
            await new Promise(resolve => {
                this.waitingQueue.push(resolve);
            });
        }
        this.locked = true;
    }

    release() {
        if (this.waitingQueue.length > 0) {
            const nextResolve = this.waitingQueue.shift();
            nextResolve();
        } else {
            this.locked = false;
        }
    }
}

export default Lock;

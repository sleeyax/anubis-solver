class Worker {
    constructor(script) {
        this.listeners = new Map();
        this.onmessage = null;
        this.onerror = null;
        this.terminated = false;

        if (script.includes("base64")) {
            const source = atob(script.replace("data:application/javascript;base64,", ""));
            this.evaluate(source);
        } else {
            throw new Error('Script must be base64 encoded');
        }
    }

    terminate() {
        this.terminated = true;
        this.listeners.clear();
        // Force an error to break the infinite loop
        // this._handleError(new Error("Worker terminated"));
    }

    evaluate(script) {
        const workerScope = {
            self: this,
            postMessage: this.postMessageFromWorker.bind(this),
            addEventListener: this.addEventListener.bind(this),
        };

        (new Function('workerScope', `
            with(workerScope) {
                ${script}
            }
        `))(workerScope);
    }

    postMessage(data) {
        if (this.terminated) return;
        const event = { data };
        const listeners = this.listeners.get('message') || [];
        listeners.forEach(listener => listener(event));
    }

    postMessageFromWorker(message) {
        if (this.onmessage) {
            this.onmessage({data: message});
        }
    }

    _handleError(error) {
        const event = { error, message: error.message };
        const listeners = this.listeners.get('error') || [];
        listeners.forEach(listener => listener(event));
        if (this.onerror) {
            this.onerror(event);
        }
    }

    addEventListener(type, listener) {
        if (!this.listeners.has(type)) {
            this.listeners.set(type, []);
        }
        this.listeners.get(type).push(listener);
    }
}
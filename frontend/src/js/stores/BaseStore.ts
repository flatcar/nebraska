import EventEmitter from 'events';

var CHANGE_EVENT = 'change';

class Store extends EventEmitter {

  constructor() {
    super();
    this.setMaxListeners(100);
  }

  emitChange() {
    this.emit(CHANGE_EVENT);
  }

  addChangeListener(callback: () => void) {
    this.on(CHANGE_EVENT, callback);
  }

  removeChangeListener(callback: () => void) {
    this.removeListener(CHANGE_EVENT, callback);
  }
}

export default Store;

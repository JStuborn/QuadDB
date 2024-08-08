function ensureInstance(obj, Class) {
    if (!(obj instanceof Class)) {
        throw new TypeError("Cannot call a class as a function");
    }
}

function defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
        var prop = props[i];
        prop.enumerable = prop.enumerable || false;
        prop.configurable = true;
        if ("value" in prop) prop.writable = true;
        Object.defineProperty(target, prop.key, prop);
    }
}

function applyDecorators(Constructor, protoDecorators, staticDecorators) {
    if (protoDecorators) defineProperties(Constructor.prototype, protoDecorators);
    if (staticDecorators) defineProperties(Constructor, staticDecorators);
    return Constructor;
}

(function () {
    var defineProperty = Object.defineProperty;
    var setName = function (target, name) {
        return defineProperty(target, "name", {
            value: name,
            configurable: true,
        });
    };

    var Notification = function () {
        function Notification(options) {
            ensureInstance(this, Notification);

            this.backgroundColor = options.backgroundColor || '#ffffff';
            this.title = options.title || '';
            this.text = options.text || '';
            this.speed = options.speed || 500;
            this.position = options.position || 'right top';
            this.autoclose = options.autoclose !== undefined ? options.autoclose : true;
            this.autotimeout = options.autotimeout || 3000;

            if (!this.checkRequirements()) {
                console.error("You must specify 'title' or 'text' at least.");
                return;
            }

            this.setContainer();
            this.setWrapper();
            this.setPosition();
            this.setContent();
            this.container.prepend(this.wrapper);
            this.setEffect();
            this.notifyIn(this.selectedNotifyInEffect);
            if (this.autoclose) this.autoClose();
        }

        applyDecorators(Notification, [{
            key: "checkRequirements",
            value: function checkRequirements() {
                return !!(this.title || this.text);
            }
        }, {
            key: "setContainer",
            value: function setContainer() {
                var container = document.querySelector(".sn-notifications-container");
                if (!container) {
                    this.container = document.createElement("div");
                    this.container.classList.add("sn-notifications-container");
                    document.body.appendChild(this.container);
                } else {
                    this.container = container;
                }
            }
        }, {
            key: "setPosition",
            value: function setPosition() {
                var classList = this.container.classList;
                classList.remove('sn-is-center', 'sn-is-left', 'sn-is-right', 'sn-is-top', 'sn-is-bottom', 'sn-is-x-center', 'sn-is-y-center');
            
                if (this.position === 'center') {
                    classList.add('sn-is-center');
                }
                if (this.position.includes('left')) {
                    classList.add('sn-is-left');
                }
                if (this.position.includes('right')) {
                    classList.add('sn-is-right');
                }
                if (this.position.includes('top')) {
                    classList.add('sn-is-top');
                }
                if (this.position.includes('bottom')) {
                    classList.add('sn-is-bottom');
                }
                if (this.position.includes('x-center')) {
                    classList.add('sn-is-x-center');
                }
                if (this.position.includes('y-center')) {
                    classList.add('sn-is-y-center');
                }
            }
            }, {
            key: "setWrapper",
            value: function setWrapper() {
                this.wrapper = document.createElement("div");
                this.wrapper.classList.add("sn-notify");
                this.wrapper.style.backgroundColor = this.backgroundColor;
                this.wrapper.style.transitionDuration = this.speed + "ms";
                this.wrapper.innerHTML = "<div class=\"sn-notify-content\">" + (this.title ? "<div class=\"sn-notify-title\">" + this.title.trim() + "</div>" : "") + (this.text ? "<div class=\"sn-notify-text text-white\">" + this.text.trim() + "</div>" : "") + "</div>";

                this.container.prepend(this.wrapper);
            }
        }, {
            key: "setContent",
            value: function setContent() {
                // No additional content setup needed
            }
        }, {
            key: "setEffect",
            value: function setEffect() {
                this.selectedNotifyInEffect = this.fadeIn;
                this.selectedNotifyOutEffect = this.fadeOut;
            }
        }, {
            key: "notifyIn",
            value: function notifyIn(effect) {
                var _this = this;
                setTimeout(function () {
                    _this.wrapper.classList.add('sn-notify-fade', 'sn-notify-fade-in');
                }, 100);
            }
        }, {
            key: "autoClose",
            value: function autoClose() {
                var _this2 = this;
                setTimeout(function () {
                    _this2.wrapper.classList.remove('sn-notify-fade-in');
                    setTimeout(function () {
                        _this2.wrapper.remove();
                    }, _this2.speed);
                }, this.autotimeout + this.speed);
            }
        }]);

        return Notification;
    }();

    setName(Notification, "Notify");

    globalThis.Notify = Notification;
})();

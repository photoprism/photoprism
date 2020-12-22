/*
Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

Author: Kay-Uwe (Kiwi) Lorenz <tabkiwi@gmail.com>
*/

// keymap is a mapping of "context" => "keys"
//
// key action can be:
// string:
//    <handler>.path.to.function
// 
const keymap = {
    'any': {
        'Escape': 'app.$el.focus'
    },
    'app': {},
    'navigation': {
        'ctrl+\\': 'navigation.toggleSidebar',
        'g b': {route: '/photos', continue: true},
        'g b m': '/photos/monochrome',
        'g b p': '/photos/panoramas',
        'g b t': '/photos/stacks',
        'g b s': '/photos/scans',
        //'g,b,s': 'review',
        //'g,b,s': 'review',
        'g b a': '/archive',
        'g f': '/favorites',
        'g v': '/videos',
        'g a': {route: '/albums', continue: true},
        'g a u': '/unsorted',
        'g F': '/folders',
        'g P': '/private',
        'g c': '/calendar',
        'g m': '/moments',
        'g p': '/places',
        'g p s': {route: '/states', continue: true},
        'g l': '/labels',
        'g L': {route: '/library', continue: true},
        'g L o': '/library/files',
        'g L h': '/library/hidden',
        'g L e': '/library/errors',
        'g h': 'navigation.goHome',
        'g S': {route: '/settings', continue: true},
        'g S a': '/about',
        'g S f': '/feedback',
        'g S l': '/about/license',
    },
    'toolbar': {
        '/': 'toolbar.$refs.search.focus',
        'v l': {name: 'toolbar.setView', args: 'list'},
        'v m': {name: 'toolbar.setView', args: 'mosaic'},
        'v c': {name: 'toolbar.setView', args: 'cards'},
    },
    'photo': {
        'e': 'photo.edit',
        'f': 'photo.activePhoto.toggleLike',
        'p': 'photo.activePhoto.togglePrivate',
        'h': 'photo.prev',
        'j': 'photo.down',
        'k': 'photo.up',
        'l': 'photo.next',
        'o': 'photo.open',
        'Enter': 'photo.open',
        'e': 'photo.edit',
        'ArrowRight': 'photo.next',
        'ArrowUp': 'photo.up',
        'ArrowDown': 'photo.down',
        'ArrowLeft': 'photo.prev',
        'Space': 'photo.toggleSelection',
        //'ArrowDown': 
    },

    'viewer': {
        'Escape': {name: 'viewer.$refs.close.click', resetFocus: true},

        // ondownload only if viewer.config.settings.features.download
        'd': 'viewer.onDownload',
        'e': 'viewer.onEdit',
        'f': 'viewer.$refs.like.click',

        'F': 'viewer.$refs.fullscreen.click',
        'z': 'viewer.$refs.zoom.click',
        'p': 'viewer.onSlideshow',

        'h': 'viewer.$refs.prev.click',
        'ArrowLeft': 'viewer.$refs.prev.click',
        'j': 'viewer.$refs.next.click',

        'k': 'viewer.$refs.prev.click',
        'l': 'viewer.$refs.next.click',
        'ArrowRight': 'viewer.$refs.next.click',

        'Space': 'viewer.onPlay',
    }
}

export default class {
    constructor() {
        this.currentKey = '';
        this.app = null;
        console.log("keymap", keymap);
        this.setKeymap(keymap);
        this.contexts = {};
        this.contextStack = [];
        this.keySequenceTimeout = 700;
        this.activeIndex = -1;
    }

    activate(object, name) {
        this.contexts[name] = object;
        this.contextStack.push(name)
    }

    deactivate(name) {
        delete this.contexts[name];
        const index = this.contextStack.indexOf(name);
        if (index > -1) {
            this.contextStack.splice(index, 1)
        }
    }

    isHidden(el) {
        var style = window.getComputedStyle(el);
        return (style.display === 'none')
    }

    setKeymap(keymap) {
        // normalize keymap
        this.keymap = {};

        for (var ctx in keymap) {
            this.keymap[ctx] = {};

            for (var key in keymap[ctx]) {
                let action = keymap[ctx][key];

                if (typeof action === 'function') {
                    this.keymap[ctx][key] = action;
                } else {
                    // clone object (incl. strings, etc.)
                    action = JSON.parse(JSON.stringify(action));

                    if (typeof action === 'string') {
                        if (action.match(/^\//)) {
                            action = { route: action };
                        } else {
                            let path = action.split(".");
                            action = {};
                            action.command = path[0];
                            action.name = path.splice(1);
                        }
                    }

                    if (typeof action.name === 'string') {
                        action.name = action.name.split(".")
                        if (!action.command) {
                            action.command = action.name[0];
                            action.name = action.name.splice(1);
                        }
                    }

                    if (typeof action.object === 'string') {
                        action.object = action.object.split(".");
                    }
                }
                this.keymap[ctx][key] = action;
            }
        }
    }

    setApp(app) {
        this.app = app;
        this.activate(app, 'app');
    }

    keyAction(action, event) {
        if (!action) return;
        if (!this.app) return;

        if (typeof action === 'function') {
            action.apply(this, event);
        } else if (action.route) {
            this.app.$router.push(action.route).catch(error => {
                if (error.name != "NavigationDuplicated") {
                    throw error;
                }
            });
        } else {
            var context = this.contexts[action.command]
            console.log("command", action.command, "context", context, "name", action.name);
            var obj;

            for (var i=0; i<action.name.length; i++) {
                obj = context;

                if (!(action.name[i] in context)) {
                    console.log("context", context, "has no attribute", action.name[i]);
                    return;
                }

                context = context[action.name[i]];
            }
            if (action.object) {
                var obj = this.contexts[action.object[0]]
                for (var i=1; i < action.object.length; i++) {
                    obj = obj[action.object[i]];
                }
            }
            var args = action.args || [];
            if (!(args instanceof Array)) {
                args = [ args ]
            }

            console.log("apply", obj, context, args);

            if (action.argEvent === true) {
                args = [ event ] + args
            }

            context.apply(obj, args);
        }

        if (action.resetFocus) {
            this.app.$nextTick(() => {
                this.app.$el.focus();
            });
        }
    }

    handleKey(event, context) {
        console.log("handleKey", event, "context", context);

        if (!event.key) return;

        setTimeout(() => {
            console.log("stop key collecting, dispose:", this.currentKey)
            this.currentKey = ''
        }, this.keySequenceTimeout);

        let key = event.key;
        if (key == ' ') {
            key = 'Space';
        }

        if (key == "Alt" || key == "Control" || key == "Shift" || key == "Meta") {
            return;
        }

        if (event.ctrlKey) {
            key = 'ctrl+'+key
        }

        if (event.altKey) {
            key = 'alt+'+key
        }

        if (event.metaKey) {
            key = 'meta+'+key
        }

        if (this.currentKey) {
            this.currentKey += ' '+key;
        } else {
            this.currentKey = key
        }

        //this.app.$notify.info(' '+this.currentKey, this.keySequenceTimeout);

        let action = null;

        console.log("lookup", this.currentKey)
        console.log("contexts", this.contexts)

        for (var i=this.contextStack.length-1; i >= 0; i--)
        {
            let ctx = this.contextStack[i]
            console.log("check ctx", ctx)
            if (!(ctx in this.keymap)) {
                console.log("ctx not in keymap:", ctx);
                continue;
            }
            if (this.isHidden(this.contexts[ctx].$el)) {
                console.log("ctx element hidden:", ctx);
                continue;
            }
            console.log("check if", this.currentKey, "is in", this.keymap)
            if (this.currentKey in this.keymap[ctx]) {
                action = this.keymap[ctx][this.currentKey]
            }
        }
        console.log("action", action)

        if (action === null && this.currentKey in this.keymap.any) {
            action = this.keymap.any[this.currentKey];
        }
        console.log("action", action)

        if (action !== null) {
            this.keyAction(action, event);
            if (action.continue !== true) {
                this.currentKey = '';
            }
            event.stopPropagation();
            event.preventDefault();
        }
    }
}
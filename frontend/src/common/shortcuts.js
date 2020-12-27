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

// keymap is a mapping of "context" => ""
//
// key action can be:
// string:
//    <handler>.path.to.function
// 

const keymap = [
    {
      section: "Navigation",
      context: 'navigation',
      actions: [
        {
            name: "Display shortcuts help",
            context: 'any',
            keys: '?',
            action: 'navigation.showShortcutshelp'
        },
        {
            name: 'Toggle sidebar', 
            keys: 'ctrl+\\', 
            action: 'navigation.toggleSidebar', 
        },
        {
            name: 'Go to browse photos', 
            keys: 'g b', 
            action: {route: '/photos', continue: true}, 
        },
        { 
            name: 'Go to browse monochromes', 
            keys: 'g b m', 
            action: '/photos/monochrome', 
        },
        {
            name: 'Go to browse panoramas', 
            keys: 'g b p', 
            action: '/photos/panoramas', 
        },
        {
            name: 'Go to browse stacks', 
            keys: 'g b t', 
            action: '/photos/stacks', 
        },
        {
            name: 'Go to browse scans', 
            keys: 'g b s', 
            action: '/photos/scans', 
        },
    //    { name: '', 
    //      keys: //'g,b,s', 
    //      action: 'review', },
    //    { name: '', 
    //      keys: //'g,b,s', 
    //      action: 'review', },
        {
            name: 'Go to browse archive', 
            keys: 'g b a', action: '/archive',
        },
        {
            name: 'Go to favorites', 
            keys: 'g f', 
            action: '/favorites', 
        },
        {
            name: 'Go to videos', 
            keys: 'g v', 
            action: '/videos', 
        },
        {
            name: 'Go to albums', 
            keys: 'g a', 
            action: {route: '/albums', continue: true}, 
        },
        {
            name: 'Go to unsorted photos (not in album)', 
            keys: 'g a u', 
            action: '/unsorted', 
        },
        {
            name: 'Go to folders', 
            keys: 'g F', 
            action: '/folders', 
        },
        {
            name: 'Go to private', 
            keys: 'g P', 
            action: '/private', 
        },
        {
            name: 'Go to calendar', 
            keys: 'g c', 
            action: '/calendar', 
        },
        {
            name: 'Go to moments', 
            keys: 'g m', 
            action: '/moments', 
        },
        {
            name: 'Go to places', 
            keys: 'g p', 
            action: '/places', 
        },
        {
            name: 'Go to states', 
            keys: 'g p s', 
            action: {route: '/states', continue: true}, 
        },
        {
            name: 'Go to label management', 
            keys: 'g l', 
            action: '/labels', 
        },
        {
            name: 'Go to library', 
            keys: 'g L', 
            action: {route: '/library', continue: true}, 
        },
        {
            name: 'Go to library/files', 
            keys: 'g L o', 
            action: '/library/files', 
        },
        {
            name: 'Go to hidden photos', 
            keys: 'g L h', 
            action: '/library/hidden', 
        },
        {
            name: 'Go to import/index error view', 
            keys: 'g L e', 
            action: '/library/errors', 
        },
        {
            name: 'Go home', 
            keys: 'g h', 
            action: 'navigation.goHome', 
        },
        {
            name: 'Go to settings', 
            keys: 'g S', 
            action: {route: '/settings', continue: true}, 
        },
        {
            name: 'Go to about page', 
            keys: 'g S a', 
            action: '/about', 
        },
        {
            name: 'Go to feedback page', 
            keys: 'g S f', 
            action: '/feedback', 
        },
        {
            name: 'Go to license page', 
            keys: 'g S l', 
            action: '/about/license', 
        },
        ]
    },
    {
        section: "Photo browser",
        context: 'photo',
        actions: [
            {
                name: 'Edit current photo',
                keys: 'e',
                action: 'photo.edit',
            },
            {
                name: 'Toggle favorite (like)',
                keys: 'f',
                action: 'photo.activePhoto.toggleLike',
            },
            {
                name: 'Toggle private',
                keys: 'p',
                action: 'photo.activePhoto.togglePrivate',
            },
            {
                name: 'Navigate to previous (left) photo',
                action: 'photo.prev',
                keys: ['ArrowLeft', 'h' ],
            },
            {
                name: 'Navigate a photo down',
                action: 'photo.down',
                keys: ['ArrowDown', 'j' ],
            },
            {
                name: 'Navigate a photo up',
                action: 'photo.up',
                keys: ['ArrowUp', 'k'],
            },
            {
                action: 'photo.next',
                name: 'Naviate to next (right) photo',
                keys: ['ArrowRight', 'l'],
            },
            {
                name: 'Open the current photo',
                action: 'photo.open',
                keys: ['Enter', 'o'],
            },
            {
                name: 'Toggle selection of current photo',
                keys: 'Space',
                action: 'photo.toggleSelection',
            },
        ],
    },
    {
        section: "Photo viewer",
        context: 'viewer',
        actions: [
            {
                name: 'Close photo viewer',
                keys: 'Escape',
                action: {name: 'viewer.$refs.close.click', resetFocus: true},
            },
            {
                name: 'Download current photo',
                keys: 'd',
                action: 'viewer.onDownload',
            },
            {
                name: 'Edit current photo',
                keys: 'e',
                action: 'viewer.onEdit',
            },
            {
                name: 'Toggle favorite (like)',
                keys: 'f',
                action: 'viewer.$refs.like.click',
            },
            {
                name: 'Toggle selection',
                keys: 'Space',
                action: 'viewer.$refs.select.click',
            },
            {
                name: 'Toggle fullscreen view',
                keys: 'F',
                action: 'viewer.$refs.fullscreen.click',
            },
            {
                name: 'Toggle zoom view',
                keys: 'z',
                action: 'viewer.$refs.zoom.click',
            },
            {
                name: 'Start (play) slideshow',
                keys: 'p',
                action: 'viewer.onSlideshow',
            },
            {
                name: 'Previous photo',
                keys: ['ArrowLeft', 'h', 'ArrowUp', 'k'],
                action: 'viewer.$refs.prev.click',
            },
            {
                name: 'Next photo',
                keys: ['ArrowRight', 'l', 'ArrowDown', 'j'],
                action: 'viewer.$refs.next.click',
            },
            {
                name: 'Play video',
                keys: 'Enter',
                action: 'viewer.onPlay',
            },
        ]
    },
    // does not work :(
    // {
    //     section: "Video player",
    //     context: 'player',
    //     actions: [
    //         { 
    //             name: 'Toggle play',
    //             keys: 'Space',
    //             action: 'player.$refs.player.togglePlay'
    //         },
    //         {
    //             name: "Close Player",
    //             keys: "Escape",
    //             action: "player.onClose"
    //         }
    //     ]
    // },
    {
        section: "Toolbar actions",
        context: 'toolbar',
        actions: [
            { 
                name: 'Search a photo',
                keys: '/',
                action: 'toolbar.$refs.search.focus',
            },
            { 
                name: 'Set list view mode',
                keys: 'v l',
                action: {name: 'toolbar.setView', args: 'list'},
            },
            { 
                name: 'Set mosaic view mode',
                keys: 'v m',
                action: {name: 'toolbar.setView', args: 'mosaic'},
            },
            { 
                name: 'Set card view mode',
                keys: 'v c',
                action: {name: 'toolbar.setView', args: 'cards'},
            },

        ]
    },
    {
        section: "Form elements",
        context: 'any',
        actions: [
            {
                name: 'Escape current focussed element',
                context: 'any',
                keys: 'Escape',
                action: 'app.$el.focus'
            },
        ]
    },
    {
        section: "Dialog",
        context: 'dialog',
        actions: [
            {
                name: "Cancel dialog",
                keys: 'Escape',
                action: function() {
                    if (this.contexts['dialog'].close) {
                        this.contexts['dialog'].close()
                    } else if (this.contexts['dialog'].cancel) {
                        this.contexts['dialog'].cancel()
                    }
                }
            },
            {
                name: "Confirm dialog",
                keys: 'ctrl+Enter',
                action: 'dialog.confirm'
            },
        ]
    },
    
]
        // ondownload only if viewer.config.settings.features.download

const displayKeyMap = {
    'ArrowLeft': '←',
    'ArrowUp': "↑",
    'ArrowRight': '→',
    'ArrowDown': '↓',
    'Escape': 'Esc',
    'Enter': '⏎',
    'Shift': '⇧',
}

export default class {

    constructor() {
        this.currentKey = '';
        this.app = null;
        this.actions = keymap;
        this.setKeymap(keymap);
        this.contexts = {};
        this.contextStack = [];
        this.keySequenceTimeout = 700;
        this.activeIndex = -1;
    }

    get debug() {
        if (this.app) {
            return this.app.$config.debug
        } else {
            return true
        }
    }

    log_debug() {
        if (this.debug) {
            const args = [].concat([ "Shortcuts -" ], Array.from(arguments))
            console.debug.apply(console, args)
        }
    }

    log_info() {
        if (this.debug) {
            const args = [].concat([ "Shortcuts -" ], Array.from(arguments))
            console.info.apply(console, args)
        }
    }

    activate(object, name) {
        this.log_debug("Activate", name)

        this.contexts[name] = object;
        this.contextStack.push(name)
        this.resetFocus();
    }

    deactivate(name) {
        this.log_debug("Deactivate", name)

        delete this.contexts[name];
        const index = this.contextStack.indexOf(name);
        if (index > -1) {
            this.contextStack.splice(index, 1)
        }
        this.resetFocus();
    }

    isHidden(el) {
        var style = window.getComputedStyle(el);
        return (style.display === 'none')
    }

    setKeymap(keymap) {
        // normalize keymap
        this.keymap = {};

        let shortcuts = this

        keymap.forEach( (spec) => {
            let defaultContext = spec.context
            spec.actions.forEach( (keySpec) => {
                let ctx = keySpec.context || defaultContext
                let keys = keySpec.keys
                if (!(keys instanceof Array)) {
                    keys = [ keys ]
                }
                keySpec.displayKeys = []

                let action = keySpec.action

                if (!shortcuts.keymap[ctx]) {
                    shortcuts.keymap[ctx] = {}
                }

                keys.forEach( (key) => {
                    let displayKey = []
                    key.split(/\s+/).forEach( (k) => {
                        let _modifiers = k.split(/\+/)
                        let _key = _modifiers[_modifiers.length-1]
                        _modifiers = _modifiers.slice(0, -1)
                        if (_key.match(/^\w+$/)) {
                            if (_key == _key.toUpperCase()) {
                                _modifiers.push('Shift')
                                _key = _key.toLowerCase()
                            }
                        }

                        displayKey.push({
                            key: displayKeyMap[_key] || _key,
                            modifiers: _modifiers.map((m) => {
                                let modifier = m.charAt(0).toUpperCase() + m.slice(1).toLowerCase()
                                return displayKeyMap[modifier] || modifier
                            })
                        })
                    })

                    this.log_debug("displayKey", displayKey)

                    keySpec.displayKeys.push(displayKey)

                    if (typeof action === 'function') {
                        shortcuts.keymap[ctx][key] = action;
                        return
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
                    shortcuts.keymap[ctx][key] = action;
                })
            })
        })
/*
        for (var ctx in keymap) {
            this.keymap[ctx] = {};

            for (var key in keymap[ctx]) {
                let action = keymap[ctx][key];

            }
        }
   */ 
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

            this.log_debug("command", action.command, "context", context, "name", action.name);

            var obj;

            for (var i=0; i<action.name.length; i++) {
                obj = context;

                if (!(action.name[i] in context)) {
                    this.log_debug("context", context, "has no attribute", action.name[i]);
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

            this.log_debug("apply", obj, context, args);

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

    resetFocus() {
        if (this.app && this.app.$el) {
            this.app.$el.focus()
        }
    }

    handleKey(event, context) {
        if (!event.key) return;

        this.log_debug("Handle Key --->", event.key, "event", event, "context", context);

        setTimeout(() => {
            this.log_debug("stop key collecting, dispose:", this.currentKey)
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

        this.log_debug("lookup", this.currentKey)
        this.log_debug("contexts", this.contexts)

        for (var i=this.contextStack.length-1; i >= 0; i--)
        {
            let ctx = this.contextStack[i]
            this.log_debug("check ctx", ctx)

            if (!(ctx in this.keymap)) {
                this.log_debug("ctx not in keymap:", ctx);
                continue;
            }
            if (this.isHidden(this.contexts[ctx].$el)) {
                this.log_debug("ctx element hidden:", ctx);
                continue;
            }
            this.log_debug("check if", this.currentKey, "is in", this.keymap[ctx])

            if (this.currentKey in this.keymap[ctx]) {
                action = this.keymap[ctx][this.currentKey]
                break;
            }
        }
        this.log_debug("action so far", action)

        if (action === null && this.currentKey in this.keymap.any) {
            action = this.keymap.any[this.currentKey];
        }

        this.log_debug("found action", action)

        if (action !== null) {
            this.keyAction(action, event);
            if (action.continue !== true) {
                this.currentKey = '';
            }
            event.stopPropagation();
            event.preventDefault();
        }
        this.log_debug("Done Key --->", event.key)
    }
}
// CodeMirror, copyright (c) by Marijn Haverbeke and others
// Distributed under an MIT license: http://codemirror.net/LICENSE

// A rough approximation of Sublime Text's keybindings
// Depends on addon/search/searchcursor.js and optionally addon/dialog/dialogs.js

(function(mod) {
    mod(CodeMirror);
})(function(CodeMirror) {
  "use strict";

  var keyMap = CodeMirror.keyMap;

  keyMap.leeio = {
    "Shift-Tab": "indentLess",
    "Ctrl-F": "findPersistent",
    "Ctrl-H": "replaceAll",
    "F3": "findNext",
    "Shift-F3": "findPrev",
    "fallthrough": "default"
  };

  CodeMirror.normalizeKeyMap(keyMap.leeio);
});

//-- Toastr --//

toastr.options = {
  "positionClass": "toast-bottom-right",
  "newestOnTop": false
}

//-- Editor --//

var pasteContent = document.getElementById("paste_content");

var editor = CodeMirror.fromTextArea(pasteContent, {
    lineNumbers: true,
    indentUnit: 4,
    matchBrackets: true,
    // highlightSelectionMatches: {showToken: /\w/, annotateScrollbar: true},
    // smartIndent: false,
    keyMap: "leeio",
    inputStyle: "textarea",
    mode: "text/plain"
});

editor.setOption("fullScreen", true);
editor.focus();


//-- Editor Shortcuts --//

function setDarkMode() {
    var x = document.getElementsByClassName("CodeMirror");
    for (var i = 0; i < x.length; i++) {
        addClass(x[i], "CodeMirror-dark");
    }
}

if (HashParameters.get("dark") !== undefined) {
    setDarkMode();
    document.cookie = "dark=1";
}
if (HashParameters.get("light") !== undefined) {
    document.cookie = "dark= ; expires = Thu, 01 Jan 1970 00:00:00 GMT";
}


if (document.cookie.split(';').filter((item) => item.includes('dark=1')).length) {
    setDarkMode();
}

if (window.matchMedia) {
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        setDarkMode();
    }
}

var wordWrap = false;
document.addEventListener('keydown', function (e) {
    if (e.altKey && e.keyCode == 90) {
        e.preventDefault();
        wordWrap = (wordWrap == true) ? false : true;
        editor.setOption("lineWrapping", wordWrap);
    }
});



//-- Navbar --//

var p = document.getElementById("navbar");
setBodyPadding();

function setBodyPadding() {
    var style = p.currentStyle || window.getComputedStyle(p);
    document.getElementsByClassName("CodeMirror-fullscreen")[0].style.marginTop = (parseInt(style.height, 10))+'px';
}

window.onresize = function(e) {
    setBodyPadding();
};

var btnToggleNav = document.getElementById("btn-togglenav");
var divNavInputToggle = document.getElementById("nav-input-toggle");

btnToggleNav.onclick = function() {
    if (hasClass(divNavInputToggle, 'collapsed')) {
        removeClass(divNavInputToggle, 'collapsed');
    } else {
        addClass(divNavInputToggle, 'collapsed');
    }

    setBodyPadding();
}


//-- Handle syntax highlighting selection --//

var inputSyntax = document.getElementById("paste_syntax");
var inputExpires = document.getElementById("paste_expires");

function getSyntax() {
    return inputSyntax.options[inputSyntax.selectedIndex].value;
}

function setSyntax(val) {
    setSelect(inputSyntax, val);
}

function getExpires() {
    return inputExpires.options[inputExpires.selectedIndex].value;
}

function setExpires(val) {
    setSelect(inputExpires, val);
}

function setEditorSyntax() {
    editor.setOption("mode", getSyntax());
}

inputSyntax.onchange = function() {
    setEditorSyntax();
    setPrettifyButton();
}


//-- Form handling --//

var btnViewRaw = document.getElementById("btn-view-raw");
var btnViewRawLink = document.getElementById("btn-view-raw-link");

function addPaste() {
    PasswordGenerator.length = 24;
    PasswordGenerator.symbols = false;
    var encryptionKey = PasswordGenerator.generate();
    var encryptedText = '' + CryptoJS.AES.encrypt(editor.getValue(), encryptionKey);

    var syntax = getSyntax();
    var expires = getExpires();

    $.ajax({
        url: '/api/v1/paste',
        timeout: 90000,
        type: 'POST',
        dataType: "json",
        contentType: 'application/json',
        data: JSON.stringify({
            content: encryptedText,
            syntax: syntax,
            expires: parseInt(expires)
        }),
        error: function(jqXHR, textStatus, errorThrown) {            
            var msg = jqXHR.responseText;
            try {
                msg = JSON.parse(msg);
            } catch(e) {}

            toastr.error(msg, null, {timeOut: 2000, extendedTimeOut: 1000});
        },
        success: function(data, textStatus, jqXHR) {
            history.pushState(null, null, "/"+data.id);
            HashParameters.set("encryptionKey", encryptionKey);

            btnViewRawLink.href = '/raw/'+data.id+'#encryptionKey='+encryptionKey
            removeClass(btnViewRaw, 'hidden');

            toastr.success("Paste encrypted and saved", null, {timeOut: 2000, extendedTimeOut: 1000});

            if (expires == -2) {
                toastr.warning("Burn after read active", null, {timeOut: 2000, extendedTimeOut: 1000});
            }
        }
    });
}
function invokeGetPaste() {
    var idMatch = window.location.pathname.match(/([a-zA-Z0-9-]+)/);
    if (idMatch != null) {
        getPaste(idMatch[1]);
    }
}

function getPaste(id) {
    var encryptionKey = HashParameters.get("encryptionKey");
    var infoToast = toastr.info('Retrieving paste', null);
    $.ajax({
        url: '/api/v1/paste/'+id,
        timeout: 90000,
        type: 'GET',
        dataType: "json",
        contentType: 'application/json',
        error: function(jqXHR, textStatus, errorThrown) {
            var msg = jqXHR.responseText;
            try {
                msg = JSON.parse(msg);
            } catch(e) {}

            toastr.error(msg, null, {timeOut: 2000, extendedTimeOut: 1000});
        },
        success: function(data, textStatus, jqXHR) {
            if (encryptionKey != undefined && encryptionKey != null) {
                editor.setValue('' + CryptoJS.AES.decrypt(data.content, encryptionKey).toString(CryptoJS.enc.Utf8));
            } else {
                editor.setValue(data.content);
            }

            if (data.burnt == true) {
                toastr.warning("Paste burnt", null, {timeOut: 2000, extendedTimeOut: 1000});
            }

            setSyntax(data.syntax);
            setExpires(data.expires);
            setEditorSyntax();
            invokeSetMarkers();
            setPrettifyButton();

            btnViewRawLink.href = '/raw/'+id+((encryptionKey != null) ? '#encryptionKey='+encryptionKey : '');
            removeClass(btnViewRaw, 'hidden');
        }
    });
    toastr.clear(infoToast, { force: true });
}

var btnGo = document.getElementById("btn-go");

btnGo.onclick = function() {
    addPaste();
}

document.addEventListener('keydown', function (e) {
    if (e.ctrlKey && e.keyCode == 13) {
        e.preventDefault();
        addPaste();
    }
});


//-- Prettify --//

var btnPrettify = document.getElementById("btn-prettify");

function setPrettifyButton() {
    var syntax = getSyntax();
    switch (syntax) {
        case 'application/ld+json':
            removeClass(btnPrettify, 'hidden');
            break;
        case 'text/x-yaml':
            removeClass(btnPrettify, 'hidden');
            break;
        case 'application/xml':
            removeClass(btnPrettify, 'hidden');
            break;
        case 'text/x-sql':
            removeClass(btnPrettify, 'hidden');
            break;
        case 'text/javascript':
            removeClass(btnPrettify, 'hidden');
            break;
        case 'text/css':
            removeClass(btnPrettify, 'hidden');
            break;
        default:
            addClass(btnPrettify, 'hidden');
            break;
    }
}

function doPrettify() {
    var syntax = getSyntax();
    try {
        switch (syntax) {
            case 'application/ld+json':
                prettifyJson();
                break;
            case 'text/x-yaml':
                prettifyYaml();
                break;
            case 'application/xml':
                prettifyXml();
                break;
            case 'text/x-sql':
                prettifySql();
                break;
            case 'text/javascript':
                prettifyJavascript();
                break;
            case 'text/css':
                prettifyCss();
                break;
        }
    } catch(e) {
        toastr.error('Failed to prettify: '+e.toString(), null, {timeOut: 2000, extendedTimeOut: 1000});
    }
}

function prettifyJson() {
    var val = editor.getValue();
    if (val == '') {
        val = null;
    }

    editor.setValue(JSON.stringify(JSON.parse(val), null, 2));
}

function prettifyYaml() {
    var val = editor.getValue();
    if (val == '') {
        return;
    }

    editor.setValue(YAML.stringify(YAML.parse(val), 2));
}

function prettifyXml() {
    var val = editor.getValue();
    if (val == '') {
        return;
    }

    editor.setValue(vkbeautify.xml(val, 4));
}

function prettifySql() {
    var val = editor.getValue();
    if (val == '') {
        return;
    }

    editor.setValue(sqlFormatter.format(val, {language: "sql"}));
}

function prettifyJavascript() {
    var val = editor.getValue();
    if (val == '') {
        return;
    }

    editor.setValue(js_beautify(val));
}

function prettifyCss() {
    var val = editor.getValue();
    if (val == '') {
        return;
    }

    editor.setValue(css_beautify(val));
}

btnPrettify.onclick = function() {
    doPrettify();
}


//-- Markers --//

function getMarkers() {
    var markerValue = HashParameters.get("m");

    var startMatch = null;
    var endMatch = null;

    if (markerValue != null) {
        startMatch = markerValue.match(/L(\d+)/);
        endMatch = markerValue.match(/L\d+-L(\d+)/);
    }

    return {
        start: (startMatch != null) ? startMatch[1] : null,
        end: (endMatch != null) ? endMatch[1] : null
    }
}

function jumpToLine(i) {
    editor.scrollTo(null, editor.charCoords({line: i, ch: 0}, "local").top - (editor.getScrollerElement().offsetHeight / 4) - 5);
}

function invokeSetMarkers() {
    var markers = getMarkers();
    if (markers.start != null) {

        editor.setCursor(markers.start - 1, 0);
        jumpToLine(markers.start);

        if (markers.end != null) {
            markLines(markers.start, markers.end);
        } else {
            markLines(markers.start, markers.start);
        }
    }
}

editor.on('gutterClick', function(cm, line, gutter, e) {
  if (gutter === 'CodeMirror-linenumbers') {
    var lineNumber = (line+1)
    var markers = getMarkers();
    clearTextMarker();

    if (markers.end == null && markers.start == lineNumber) {
        HashParameters.unset("m");
        return;
    }

    if (e.shiftKey) {

        // Adding/removing from selection

        if (markers.start == null || markers.start == lineNumber) {
            return setMarkedLine(lineNumber);
        }

        if (markers.start < lineNumber) {
            return setMarkedLineRange(markers.start, lineNumber);
        }

        if (markers.end != null) {
            return setMarkedLineRange(lineNumber, markers.end);
        } else {
            return setMarkedLineRange(lineNumber, markers.start);
        }

    } else {

        // New selection

       return setMarkedLine(lineNumber);
    }
  }
});

function setMarkedLine(lineNumber) {
    HashParameters.set("m", 'L'+lineNumber);
    return markLines(lineNumber, lineNumber);
}

function setMarkedLineRange(startLineNumber, endLineNumber) {
    HashParameters.set("m", 'L'+startLineNumber+'-L'+endLineNumber);
    return markLines(startLineNumber, endLineNumber)
}

function markLines(startLineNumber, endLineNumber) {
    return editor.markText(CodeMirror.Pos(parseInt(startLineNumber)-1, 0), CodeMirror.Pos(parseInt(endLineNumber), 0), {className: "cm-styledbackground"});
}

function clearTextMarker() {
    var markers = editor.getAllMarks();
    for (var i = 0; i < markers.length; i++) {
        markers[i].clear();
    }
}


//-- Helpers --//

function hasClass(el, className) {
    if (el.classList) {
        return el.classList.contains(className);
    } else {
        return !!el.className.match(new RegExp('(\\s|^)' + className + '(\\s|$)'));
    }
}

function addClass(el, className) {
    if (el.classList) {
        el.classList.add(className);
    } else {
        el.className += " " + className;
    }
}

function removeClass(el, className) {
    if (el.classList) {
        el.classList.remove(className);
    } else {
        var reg = new RegExp('(\\s|^)' + className + '(\\s|$)');
        el.className=el.className.replace(reg, ' ');
    }
}

function setSelect(select, val) {
    var opts = select.options;
    for (var opt, j = 0; opt = opts[j]; j++) {
        if (opt.value == val) {
            select.selectedIndex = j;
            break;
        }
    }
}

function initSyntax() {
    $.ajax({
        url: '/api/v1/syntax',
        timeout: 10000,
        type: 'GET',
        dataType: "json",
        contentType: 'application/json',
        error: function(jqXHR, textStatus, errorThrown) {
            var msg = jqXHR.responseText;
            try {
                msg = JSON.parse(msg);
            } catch(e) {}

            toastr.error(msg, null, {timeOut: 2000, extendedTimeOut: 1000});
        },
        success: function(data, textStatus, jqXHR) {
            for (let i=0; i < data.length; i++) {
                var option = document.createElement("option");
                option.text = data[i].label;
                option.value = data[i].syntax;
                if (data[i].default == true) {
                    option.selected = true;
                }
                inputSyntax.add(option); 
            }
        }
    });
}

function initExpires() {
    $.ajax({
        url: '/api/v1/expires',
        timeout: 10000,
        type: 'GET',
        dataType: "json",
        contentType: 'application/json',
        error: function(jqXHR, textStatus, errorThrown) {
            var msg = jqXHR.responseText;
            try {
                msg = JSON.parse(msg);
            } catch(e) {}

            toastr.error(msg, null, {timeOut: 2000, extendedTimeOut: 1000});
        },
        success: function(data, textStatus, jqXHR) {
            console.log(data)
            for (let i=0; i < data.length; i++) {
                var option = document.createElement("option");
                option.text = data[i].label;
                option.value = data[i].expires;
                if (data[i].default == true) {
                    option.selected = true;
                }
                inputExpires.add(option); 
            }
        }
    });
}

function init() {
    initSyntax()
    initExpires()
    invokeGetPaste()
}

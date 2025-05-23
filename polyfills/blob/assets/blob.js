/* Blob.js
 * A Blob, File, FileReader & URL implementation.
 * 2020-02-01
 *
 * By Eli Grey, https://eligrey.com
 * By Jimmy Wärting, https://github.com/jimmywarting
 * License: MIT
 *   See https://github.com/eligrey/Blob.js/blob/master/LICENSE.md
 *
 * Modified by Sleeyax under MIT license.
 */

function array2base64 (input) {
    var byteToCharMap = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";

    var output = [];

    for (var i = 0; i < input.length; i += 3) {
        var byte1 = input[i];
        var haveByte2 = i + 1 < input.length;
        var byte2 = haveByte2 ? input[i + 1] : 0;
        var haveByte3 = i + 2 < input.length;
        var byte3 = haveByte3 ? input[i + 2] : 0;

        var outByte1 = byte1 >> 2;
        var outByte2 = ((byte1 & 0x03) << 4) | (byte2 >> 4);
        var outByte3 = ((byte2 & 0x0F) << 2) | (byte3 >> 6);
        var outByte4 = byte3 & 0x3F;

        if (!haveByte3) {
            outByte4 = 64;

            if (!haveByte2) {
                outByte3 = 64;
            }
        }

        output.push(
            byteToCharMap[outByte1], byteToCharMap[outByte2],
            byteToCharMap[outByte3], byteToCharMap[outByte4]
        );
    }

    return output.join("");
}

(function(global) {
    (function (factory) {
        if (typeof define === "function" && define.amd) {
            // AMD. Register as an anonymous module.
            define(["exports"], factory);
        } else if (typeof exports === "object" && typeof exports.nodeName !== "string") {
            // CommonJS
            factory(exports);
        } else {
            // Browser globals
            factory(global);
        }
    })(function (exports) {
        "use strict";

        var BlobBuilder = global.BlobBuilder
            || global.WebKitBlobBuilder
            || global.MSBlobBuilder
            || global.MozBlobBuilder;

        var URL = global.URL || global.webkitURL || function (href, a) {
            a = document.createElement("a");
            a.href = href;
            return a;
        };

        var origBlob = global.Blob;
        var createObjectURL = URL.createObjectURL;
        var revokeObjectURL = URL.revokeObjectURL;
        var strTag = global.Symbol && global.Symbol.toStringTag;
        var blobSupported = false;
        var blobSupportsArrayBufferView = false;
        var blobBuilderSupported = BlobBuilder
            && BlobBuilder.prototype.append
            && BlobBuilder.prototype.getBlob;

        try {
            // Check if Blob constructor is supported
            blobSupported = new Blob(["ä"]).size === 2;

            // Check if Blob constructor supports ArrayBufferViews
            // Fails in Safari 6, so we need to map to ArrayBuffers there.
            blobSupportsArrayBufferView = new Blob([new Uint8Array([1, 2])]).size === 2;
        } catch (e) {/* eslint-disable-line no-unused-vars */}

        // Helper function that maps ArrayBufferViews to ArrayBuffers
        // Used by BlobBuilder constructor and old browsers that didn't
        // support it in the Blob constructor.
        function mapArrayBufferViews (ary) {
            return ary.map(function (chunk) {
                if (chunk.buffer instanceof ArrayBuffer) {
                    var buf = chunk.buffer;

                    // if this is a subarray, make a copy so we only
                    // include the subarray region from the underlying buffer
                    if (chunk.byteLength !== buf.byteLength) {
                        var copy = new Uint8Array(chunk.byteLength);
                        copy.set(new Uint8Array(buf, chunk.byteOffset, chunk.byteLength));
                        buf = copy.buffer;
                    }

                    return buf;
                }

                return chunk;
            });
        }

        function BlobBuilderConstructor (ary, options) {
            options = options || {};

            var bb = new BlobBuilder();
            mapArrayBufferViews(ary).forEach(function (part) {
                bb.append(part);
            });

            return options.type ? bb.getBlob(options.type) : bb.getBlob();
        }

        function BlobConstructor (ary, options) {
            return new origBlob(mapArrayBufferViews(ary), options || {});
        }

        if (global.Blob) {
            BlobBuilderConstructor.prototype = global.Blob.prototype;
            BlobConstructor.prototype = global.Blob.prototype;
        }

        // string -> buffer
        var textEncode = new TextEncoder().encode

        // buffer -> string
        var textDecode = new TextDecoder().decode

        function FakeBlobBuilder () {
            function bufferClone (buf) {
                var view = new Array(buf.byteLength);
                var array = new Uint8Array(buf);
                var i = view.length;
                while (i--) {
                    view[i] = array[i];
                }
                return view;
            }

            var create = Object.create || function (a) {
                function c () {}
                c.prototype = a;
                return new c();
            };

            function getObjectTypeName (o) {
                return Object.prototype.toString.call(o).slice(8, -1);
            }

            function isPrototypeOf(c, o) {
                return typeof c === "object" && Object.prototype.isPrototypeOf.call(c.prototype, o);
            }

            function isDataView (o) {
                return getObjectTypeName(o) === "DataView" || isPrototypeOf(global.DataView, o);
            }

            var arrayBufferClassNames = [
                "Int8Array",
                "Uint8Array",
                "Uint8ClampedArray",
                "Int16Array",
                "Uint16Array",
                "Int32Array",
                "Uint32Array",
                "Float32Array",
                "Float64Array",
                "ArrayBuffer"
            ];

            function includes(a, v) {
                return a.indexOf(v) !== -1;
            }

            function isArrayBuffer(o) {
                return includes(arrayBufferClassNames, getObjectTypeName(o)) || isPrototypeOf(global.ArrayBuffer, o);
            }

            function concatTypedarrays (chunks) {
                var size = 0;
                var j = chunks.length;
                while (j--) { size += chunks[j].length; }
                var b = new Uint8Array(size);
                var offset = 0;
                for (var i = 0; i < chunks.length; i++) {
                    var chunk = chunks[i];
                    b.set(chunk, offset);
                    offset += chunk.byteLength || chunk.length;
                }

                return b;
            }

            /********************************************************/
            /*                   Blob constructor                   */
            /********************************************************/
            function Blob (chunks, opts) {
                chunks = chunks ? chunks.slice() : [];
                opts = opts == null ? {} : opts;
                for (var i = 0, len = chunks.length; i < len; i++) {
                    var chunk = chunks[i];
                    if (chunk instanceof Blob) {
                        chunks[i] = chunk._buffer;
                    } else if (typeof chunk === "string") {
                        chunks[i] = textEncode(chunk);
                    } else if (isDataView(chunk)) {
                        chunks[i] = bufferClone(chunk.buffer);
                    } else if (isArrayBuffer(chunk)) {
                        chunks[i] = bufferClone(chunk);
                    } else {
                        chunks[i] = textEncode(String(chunk));
                    }
                }

                this._buffer = global.Uint8Array
                    ? concatTypedarrays(chunks)
                    : [].concat.apply([], chunks);
                this.size = this._buffer.length;

                this.type = opts.type || "";
                if (/[^\u0020-\u007E]/.test(this.type)) {
                    this.type = "";
                } else {
                    this.type = this.type.toLowerCase();
                }
            }
            Blob.isPolyfill = true;

            Blob.prototype.arrayBuffer = function () {
                return Promise.resolve(this._buffer.buffer || this._buffer);
            };

            Blob.prototype.text = function () {
                return Promise.resolve(textDecode(this._buffer));
            };

            Blob.prototype.slice = function (start, end, type) {
                var slice = this._buffer.slice(start || 0, end || this._buffer.length);
                return new Blob([slice], {type: type});
            };

            Blob.prototype.toString = function () {
                return "[object Blob]";
            };

            /********************************************************/
            /*                   File constructor                   */
            /********************************************************/
            function File (chunks, name, opts) {
                opts = opts || {};
                var a = Blob.call(this, chunks, opts) || this;
                a.name = name.replace(/\//g, ":");
                a.lastModifiedDate = opts.lastModified ? new Date(opts.lastModified) : new Date();
                a.lastModified = +a.lastModifiedDate;

                return a;
            }

            File.isPolyfill = true;
            File.prototype = create(Blob.prototype);
            File.prototype.constructor = File;

            if (Object.setPrototypeOf) {
                Object.setPrototypeOf(File, Blob);
            } else {
                try {
                    File.__proto__ = Blob;
                } catch (e) {/* eslint-disable-line no-unused-vars */}
            }

            File.prototype.toString = function () {
                return "[object File]";
            };

            /********************************************************/
            /*                         URL                          */
            /********************************************************/

            URL.createObjectURL = function (blob) {
                return blob instanceof Blob
                    ? "data:" + blob.type + ";base64," + array2base64(blob._buffer)
                    : createObjectURL.call(URL, blob);
            };
            URL.createObjectURL.isPolyfill = true;

            URL.revokeObjectURL = function (url) {
                revokeObjectURL && revokeObjectURL.call(URL, url);
            };
            URL.revokeObjectURL.isPolyfill = true;

            /********************************************************/
            /*                         XHR                          */
            /********************************************************/
            var _send = global.XMLHttpRequest && global.XMLHttpRequest.prototype.send;
            if (_send) {
                XMLHttpRequest.prototype.send = function (data) {
                    if (data instanceof Blob) {
                        this.setRequestHeader("Content-Type", data.type);
                        _send.call(this, textDecode(data._buffer));
                    } else {
                        _send.call(this, data);
                    }
                };
            }

            exports.Blob = Blob;
            exports.File = File;
        }

        function fixFileAndXHR () {
            var isIE = !!global.ActiveXObject || (typeof document !== "undefined" &&
                "-ms-scroll-limit" in document.documentElement.style &&
                "-ms-ime-align" in document.documentElement.style
            );

            // Monkey patched
            // IE doesn't set Content-Type header on XHR whose body is a typed Blob
            // https://developer.microsoft.com/en-us/microsoft-edge/platform/issues/6047383
            var _send = global.XMLHttpRequest && global.XMLHttpRequest.prototype.send;
            if (isIE && _send) {
                XMLHttpRequest.prototype.send = function (data) {
                    if (data instanceof Blob) {
                        this.setRequestHeader("Content-Type", data.type);
                        _send.call(this, data);
                    } else {
                        _send.call(this, data);
                    }
                };
            }

            try {
                new File([], "");
                exports.File = global.File;
            } catch (e) { // eslint-disable-line no-unused-vars
                try {
                    exports.File = new Function("class File extends Blob {" +
                        "constructor(chunks, name, opts) {" +
                        "opts = opts || {};" +
                        "super(chunks, opts || {});" +
                        "this.name = name.replace(/\\//g, \":\");" +
                        "this.lastModifiedDate = opts.lastModified ? new Date(opts.lastModified) : new Date();" +
                        "this.lastModified = +this.lastModifiedDate;" +
                        "}};" +
                        "return new File([], \"\"), File"
                    )();
                } catch (e) { // eslint-disable-line no-unused-vars
                    exports.File = function File(b, d, c) {
                        var blob = new Blob(b, c);
                        var t = c && void 0 !== c.lastModified ? new Date(c.lastModified) : new Date();

                        blob.name = d.replace(/\//g, ":");
                        blob.lastModifiedDate = t;
                        blob.lastModified = +t;
                        blob.toString = function () {
                            return "[object File]";
                        };

                        if (strTag) {
                            blob[strTag] = "File";
                        }

                        return blob;
                    };
                }
                exports.File.isPolyfill = true;
            }
        }

        if (blobSupported) {
            fixFileAndXHR();
            exports.Blob = blobSupportsArrayBufferView ? global.Blob : BlobConstructor;
        } else if (blobBuilderSupported) {
            fixFileAndXHR();
            exports.Blob = BlobBuilderConstructor;
        } else {
            FakeBlobBuilder();
        }

        /********************************************************/
        /*                FileReader constructor                */
        /********************************************************/
        function FileReader () {
            if (!(this instanceof FileReader)) {
                throw new TypeError("Failed to construct 'FileReader': Please use the 'new' operator, this DOM object constructor cannot be called as a function.");
            }

            var delegate = document.createDocumentFragment();
            this.addEventListener = delegate.addEventListener;
            this.dispatchEvent = function (evt) {
                var local = this["on" + evt.type];
                if (typeof local === "function") local(evt);
                delegate.dispatchEvent(evt);
            };
            this.removeEventListener = delegate.removeEventListener;
        }

        function _read (fr, blob, kind) {
            if (!(blob instanceof exports.Blob)) {
                throw new TypeError("Failed to execute '" + kind + "' on 'FileReader': parameter 1 is not of type 'Blob'.");
            }

            fr.result = "";

            setTimeout(function () {
                this.readyState = FileReader.LOADING;
                fr.dispatchEvent(new Event("load"));
                fr.dispatchEvent(new Event("loadend"));
            });
        }

        FileReader.EMPTY = 0;
        FileReader.LOADING = 1;
        FileReader.DONE = 2;
        FileReader.prototype.error = null;
        FileReader.prototype.onabort = null;
        FileReader.prototype.onerror = null;
        FileReader.prototype.onload = null;
        FileReader.prototype.onloadend = null;
        FileReader.prototype.onloadstart = null;
        FileReader.prototype.onprogress = null;

        FileReader.prototype.readAsDataURL = function (blob) {
            _read(this, blob, "readAsDataURL");
            this.result = "data:" + blob.type + ";base64," + array2base64(blob._buffer);
        };

        FileReader.prototype.readAsText = function (blob) {
            _read(this, blob, "readAsText");
            this.result = textDecode(blob._buffer);
        };

        FileReader.prototype.readAsArrayBuffer = function (blob) {
            _read(this, blob, "readAsText");
            // return ArrayBuffer when possible
            this.result = (blob._buffer.buffer || blob._buffer).slice();
        };

        FileReader.prototype.abort = function () {};

        exports.FileReader = global.FileReader || FileReader;
        exports.URL = global.URL || URL;

        if (strTag) {
            if (!exports.File.prototype[strTag]) exports.File.prototype[strTag] = "File";
            if (!exports.Blob.prototype[strTag]) exports.Blob.prototype[strTag] = "Blob";
            if (!exports.FileReader.prototype[strTag]) exports.FileReader.prototype[strTag] = "FileReader";
        }

        var blob = exports.Blob.prototype;
        var stream;

        try {
            new ReadableStream({ type: "bytes" });
            stream = function stream() {
                var position = 0;
                var blob = this;

                return new ReadableStream({
                    type: "bytes",
                    autoAllocateChunkSize: 524288,

                    pull: function (controller) {
                        var v = controller.byobRequest.view;
                        var chunk = blob.slice(position, position + v.byteLength);
                        return chunk.arrayBuffer()
                            .then(function (buffer) {
                                var uint8array = new Uint8Array(buffer);
                                var bytesRead = uint8array.byteLength;

                                position += bytesRead;
                                v.set(uint8array);
                                controller.byobRequest.respond(bytesRead);

                                if(position >= blob.size)
                                    controller.close();
                            });
                    }
                });
            };
        } catch (e) { // eslint-disable-line no-unused-vars
            try {
                new ReadableStream({});
                stream = function stream(blob){
                    var position = 0;

                    return new ReadableStream({
                        pull: function (controller) {
                            var chunk = blob.slice(position, position + 524288);

                            return chunk.arrayBuffer().then(function (buffer) {
                                position += buffer.byteLength;
                                var uint8array = new Uint8Array(buffer);
                                controller.enqueue(uint8array);

                                if (position == blob.size)
                                    controller.close();
                            });
                        }
                    });
                };
            } catch (e) { // eslint-disable-line no-unused-vars
                try {
                    new Response("").body.getReader().read();
                    stream = function stream() {
                        return (new Response(this)).body;
                    };
                } catch (e) { // eslint-disable-line no-unused-vars
                    stream = function stream() {
                        throw new Error("Include https://github.com/MattiasBuelens/web-streams-polyfill");
                    };
                }
            }
        }

        function promisify(obj) {
            return new Promise(function(resolve, reject) {
                obj.onload = obj.onerror = function(evt) {
                    obj.onload = obj.onerror = null;

                    evt.type === "load" ?
                        resolve(obj.result || obj) :
                        reject(new Error("Failed to read the blob/file"));
                };
            });
        }

        if (!blob.arrayBuffer) {
            blob.arrayBuffer = function arrayBuffer() {
                var fr = new exports.FileReader();
                fr.readAsArrayBuffer(this);
                return promisify(fr);
            };
        }

        if (!blob.text) {
            blob.text = function text() {
                var fr = new exports.FileReader();
                fr.readAsText(this);
                return promisify(fr);
            };
        }

        if (!blob.stream) {
            blob.stream = stream;
        }
    });
})(
    typeof self !== "undefined" && self ||
    typeof window !== "undefined" && window ||
    typeof global !== "undefined" && global ||
    this
);
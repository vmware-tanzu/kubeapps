// tslint:disable
/*eslint-disable block-scoped-var, no-redeclare, no-control-regex, no-prototype-builtins*/
"use strict";

var $protobuf = require("protobufjs/minimal");

// Common aliases
var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

$root.hapi = (function() {

    /**
     * Namespace hapi.
     * @exports hapi
     * @namespace
     */
    var hapi = {};

    hapi.release = (function() {

        /**
         * Namespace release.
         * @memberof hapi
         * @namespace
         */
        var release = {};

        release.Release = (function() {

            /**
             * Properties of a Release.
             * @memberof hapi.release
             * @interface IRelease
             * @property {string|null} [name] Release name
             * @property {hapi.release.IInfo|null} [info] Release info
             * @property {hapi.chart.IChart|null} [chart] Release chart
             * @property {hapi.chart.IConfig|null} [config] Release config
             * @property {string|null} [manifest] Release manifest
             * @property {Array.<hapi.release.IHook>|null} [hooks] Release hooks
             * @property {number|null} [version] Release version
             * @property {string|null} [namespace] Release namespace
             */

            /**
             * Constructs a new Release.
             * @memberof hapi.release
             * @classdesc Represents a Release.
             * @implements IRelease
             * @constructor
             * @param {hapi.release.IRelease=} [properties] Properties to set
             */
            function Release(properties) {
                this.hooks = [];
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Release name.
             * @member {string} name
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.name = "";

            /**
             * Release info.
             * @member {hapi.release.IInfo|null|undefined} info
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.info = null;

            /**
             * Release chart.
             * @member {hapi.chart.IChart|null|undefined} chart
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.chart = null;

            /**
             * Release config.
             * @member {hapi.chart.IConfig|null|undefined} config
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.config = null;

            /**
             * Release manifest.
             * @member {string} manifest
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.manifest = "";

            /**
             * Release hooks.
             * @member {Array.<hapi.release.IHook>} hooks
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.hooks = $util.emptyArray;

            /**
             * Release version.
             * @member {number} version
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.version = 0;

            /**
             * Release namespace.
             * @member {string} namespace
             * @memberof hapi.release.Release
             * @instance
             */
            Release.prototype.namespace = "";

            /**
             * Creates a new Release instance using the specified properties.
             * @function create
             * @memberof hapi.release.Release
             * @static
             * @param {hapi.release.IRelease=} [properties] Properties to set
             * @returns {hapi.release.Release} Release instance
             */
            Release.create = function create(properties) {
                return new Release(properties);
            };

            /**
             * Encodes the specified Release message. Does not implicitly {@link hapi.release.Release.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.Release
             * @static
             * @param {hapi.release.IRelease} message Release message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Release.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.info != null && message.hasOwnProperty("info"))
                    $root.hapi.release.Info.encode(message.info, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.chart != null && message.hasOwnProperty("chart"))
                    $root.hapi.chart.Chart.encode(message.chart, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                if (message.config != null && message.hasOwnProperty("config"))
                    $root.hapi.chart.Config.encode(message.config, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    writer.uint32(/* id 5, wireType 2 =*/42).string(message.manifest);
                if (message.hooks != null && message.hooks.length)
                    for (var i = 0; i < message.hooks.length; ++i)
                        $root.hapi.release.Hook.encode(message.hooks[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
                if (message.version != null && message.hasOwnProperty("version"))
                    writer.uint32(/* id 7, wireType 0 =*/56).int32(message.version);
                if (message.namespace != null && message.hasOwnProperty("namespace"))
                    writer.uint32(/* id 8, wireType 2 =*/66).string(message.namespace);
                return writer;
            };

            /**
             * Encodes the specified Release message, length delimited. Does not implicitly {@link hapi.release.Release.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.Release
             * @static
             * @param {hapi.release.IRelease} message Release message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Release.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Release message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.Release
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.Release} Release
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Release.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.Release();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.info = $root.hapi.release.Info.decode(reader, reader.uint32());
                        break;
                    case 3:
                        message.chart = $root.hapi.chart.Chart.decode(reader, reader.uint32());
                        break;
                    case 4:
                        message.config = $root.hapi.chart.Config.decode(reader, reader.uint32());
                        break;
                    case 5:
                        message.manifest = reader.string();
                        break;
                    case 6:
                        if (!(message.hooks && message.hooks.length))
                            message.hooks = [];
                        message.hooks.push($root.hapi.release.Hook.decode(reader, reader.uint32()));
                        break;
                    case 7:
                        message.version = reader.int32();
                        break;
                    case 8:
                        message.namespace = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Release message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.Release
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.Release} Release
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Release.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Release message.
             * @function verify
             * @memberof hapi.release.Release
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Release.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.info != null && message.hasOwnProperty("info")) {
                    var error = $root.hapi.release.Info.verify(message.info);
                    if (error)
                        return "info." + error;
                }
                if (message.chart != null && message.hasOwnProperty("chart")) {
                    var error = $root.hapi.chart.Chart.verify(message.chart);
                    if (error)
                        return "chart." + error;
                }
                if (message.config != null && message.hasOwnProperty("config")) {
                    var error = $root.hapi.chart.Config.verify(message.config);
                    if (error)
                        return "config." + error;
                }
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    if (!$util.isString(message.manifest))
                        return "manifest: string expected";
                if (message.hooks != null && message.hasOwnProperty("hooks")) {
                    if (!Array.isArray(message.hooks))
                        return "hooks: array expected";
                    for (var i = 0; i < message.hooks.length; ++i) {
                        var error = $root.hapi.release.Hook.verify(message.hooks[i]);
                        if (error)
                            return "hooks." + error;
                    }
                }
                if (message.version != null && message.hasOwnProperty("version"))
                    if (!$util.isInteger(message.version))
                        return "version: integer expected";
                if (message.namespace != null && message.hasOwnProperty("namespace"))
                    if (!$util.isString(message.namespace))
                        return "namespace: string expected";
                return null;
            };

            /**
             * Creates a Release message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.Release
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.Release} Release
             */
            Release.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.Release)
                    return object;
                var message = new $root.hapi.release.Release();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.info != null) {
                    if (typeof object.info !== "object")
                        throw TypeError(".hapi.release.Release.info: object expected");
                    message.info = $root.hapi.release.Info.fromObject(object.info);
                }
                if (object.chart != null) {
                    if (typeof object.chart !== "object")
                        throw TypeError(".hapi.release.Release.chart: object expected");
                    message.chart = $root.hapi.chart.Chart.fromObject(object.chart);
                }
                if (object.config != null) {
                    if (typeof object.config !== "object")
                        throw TypeError(".hapi.release.Release.config: object expected");
                    message.config = $root.hapi.chart.Config.fromObject(object.config);
                }
                if (object.manifest != null)
                    message.manifest = String(object.manifest);
                if (object.hooks) {
                    if (!Array.isArray(object.hooks))
                        throw TypeError(".hapi.release.Release.hooks: array expected");
                    message.hooks = [];
                    for (var i = 0; i < object.hooks.length; ++i) {
                        if (typeof object.hooks[i] !== "object")
                            throw TypeError(".hapi.release.Release.hooks: object expected");
                        message.hooks[i] = $root.hapi.release.Hook.fromObject(object.hooks[i]);
                    }
                }
                if (object.version != null)
                    message.version = object.version | 0;
                if (object.namespace != null)
                    message.namespace = String(object.namespace);
                return message;
            };

            /**
             * Creates a plain object from a Release message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.Release
             * @static
             * @param {hapi.release.Release} message Release
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Release.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults)
                    object.hooks = [];
                if (options.defaults) {
                    object.name = "";
                    object.info = null;
                    object.chart = null;
                    object.config = null;
                    object.manifest = "";
                    object.version = 0;
                    object.namespace = "";
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.info != null && message.hasOwnProperty("info"))
                    object.info = $root.hapi.release.Info.toObject(message.info, options);
                if (message.chart != null && message.hasOwnProperty("chart"))
                    object.chart = $root.hapi.chart.Chart.toObject(message.chart, options);
                if (message.config != null && message.hasOwnProperty("config"))
                    object.config = $root.hapi.chart.Config.toObject(message.config, options);
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    object.manifest = message.manifest;
                if (message.hooks && message.hooks.length) {
                    object.hooks = [];
                    for (var j = 0; j < message.hooks.length; ++j)
                        object.hooks[j] = $root.hapi.release.Hook.toObject(message.hooks[j], options);
                }
                if (message.version != null && message.hasOwnProperty("version"))
                    object.version = message.version;
                if (message.namespace != null && message.hasOwnProperty("namespace"))
                    object.namespace = message.namespace;
                return object;
            };

            /**
             * Converts this Release to JSON.
             * @function toJSON
             * @memberof hapi.release.Release
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Release.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Release;
        })();

        release.Hook = (function() {

            /**
             * Properties of a Hook.
             * @memberof hapi.release
             * @interface IHook
             * @property {string|null} [name] Hook name
             * @property {string|null} [kind] Hook kind
             * @property {string|null} [path] Hook path
             * @property {string|null} [manifest] Hook manifest
             * @property {Array.<hapi.release.Hook.Event>|null} [events] Hook events
             * @property {google.protobuf.ITimestamp|null} [lastRun] Hook lastRun
             * @property {number|null} [weight] Hook weight
             * @property {Array.<hapi.release.Hook.DeletePolicy>|null} [deletePolicies] Hook deletePolicies
             */

            /**
             * Constructs a new Hook.
             * @memberof hapi.release
             * @classdesc Represents a Hook.
             * @implements IHook
             * @constructor
             * @param {hapi.release.IHook=} [properties] Properties to set
             */
            function Hook(properties) {
                this.events = [];
                this.deletePolicies = [];
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Hook name.
             * @member {string} name
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.name = "";

            /**
             * Hook kind.
             * @member {string} kind
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.kind = "";

            /**
             * Hook path.
             * @member {string} path
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.path = "";

            /**
             * Hook manifest.
             * @member {string} manifest
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.manifest = "";

            /**
             * Hook events.
             * @member {Array.<hapi.release.Hook.Event>} events
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.events = $util.emptyArray;

            /**
             * Hook lastRun.
             * @member {google.protobuf.ITimestamp|null|undefined} lastRun
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.lastRun = null;

            /**
             * Hook weight.
             * @member {number} weight
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.weight = 0;

            /**
             * Hook deletePolicies.
             * @member {Array.<hapi.release.Hook.DeletePolicy>} deletePolicies
             * @memberof hapi.release.Hook
             * @instance
             */
            Hook.prototype.deletePolicies = $util.emptyArray;

            /**
             * Creates a new Hook instance using the specified properties.
             * @function create
             * @memberof hapi.release.Hook
             * @static
             * @param {hapi.release.IHook=} [properties] Properties to set
             * @returns {hapi.release.Hook} Hook instance
             */
            Hook.create = function create(properties) {
                return new Hook(properties);
            };

            /**
             * Encodes the specified Hook message. Does not implicitly {@link hapi.release.Hook.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.Hook
             * @static
             * @param {hapi.release.IHook} message Hook message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Hook.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.kind != null && message.hasOwnProperty("kind"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.kind);
                if (message.path != null && message.hasOwnProperty("path"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.path);
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    writer.uint32(/* id 4, wireType 2 =*/34).string(message.manifest);
                if (message.events != null && message.events.length) {
                    writer.uint32(/* id 5, wireType 2 =*/42).fork();
                    for (var i = 0; i < message.events.length; ++i)
                        writer.int32(message.events[i]);
                    writer.ldelim();
                }
                if (message.lastRun != null && message.hasOwnProperty("lastRun"))
                    $root.google.protobuf.Timestamp.encode(message.lastRun, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
                if (message.weight != null && message.hasOwnProperty("weight"))
                    writer.uint32(/* id 7, wireType 0 =*/56).int32(message.weight);
                if (message.deletePolicies != null && message.deletePolicies.length) {
                    writer.uint32(/* id 8, wireType 2 =*/66).fork();
                    for (var i = 0; i < message.deletePolicies.length; ++i)
                        writer.int32(message.deletePolicies[i]);
                    writer.ldelim();
                }
                return writer;
            };

            /**
             * Encodes the specified Hook message, length delimited. Does not implicitly {@link hapi.release.Hook.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.Hook
             * @static
             * @param {hapi.release.IHook} message Hook message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Hook.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Hook message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.Hook
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.Hook} Hook
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Hook.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.Hook();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.kind = reader.string();
                        break;
                    case 3:
                        message.path = reader.string();
                        break;
                    case 4:
                        message.manifest = reader.string();
                        break;
                    case 5:
                        if (!(message.events && message.events.length))
                            message.events = [];
                        if ((tag & 7) === 2) {
                            var end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.events.push(reader.int32());
                        } else
                            message.events.push(reader.int32());
                        break;
                    case 6:
                        message.lastRun = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 7:
                        message.weight = reader.int32();
                        break;
                    case 8:
                        if (!(message.deletePolicies && message.deletePolicies.length))
                            message.deletePolicies = [];
                        if ((tag & 7) === 2) {
                            var end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.deletePolicies.push(reader.int32());
                        } else
                            message.deletePolicies.push(reader.int32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Hook message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.Hook
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.Hook} Hook
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Hook.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Hook message.
             * @function verify
             * @memberof hapi.release.Hook
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Hook.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.kind != null && message.hasOwnProperty("kind"))
                    if (!$util.isString(message.kind))
                        return "kind: string expected";
                if (message.path != null && message.hasOwnProperty("path"))
                    if (!$util.isString(message.path))
                        return "path: string expected";
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    if (!$util.isString(message.manifest))
                        return "manifest: string expected";
                if (message.events != null && message.hasOwnProperty("events")) {
                    if (!Array.isArray(message.events))
                        return "events: array expected";
                    for (var i = 0; i < message.events.length; ++i)
                        switch (message.events[i]) {
                        default:
                            return "events: enum value[] expected";
                        case 0:
                        case 1:
                        case 2:
                        case 3:
                        case 4:
                        case 5:
                        case 6:
                        case 7:
                        case 8:
                        case 9:
                        case 10:
                            break;
                        }
                }
                if (message.lastRun != null && message.hasOwnProperty("lastRun")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.lastRun);
                    if (error)
                        return "lastRun." + error;
                }
                if (message.weight != null && message.hasOwnProperty("weight"))
                    if (!$util.isInteger(message.weight))
                        return "weight: integer expected";
                if (message.deletePolicies != null && message.hasOwnProperty("deletePolicies")) {
                    if (!Array.isArray(message.deletePolicies))
                        return "deletePolicies: array expected";
                    for (var i = 0; i < message.deletePolicies.length; ++i)
                        switch (message.deletePolicies[i]) {
                        default:
                            return "deletePolicies: enum value[] expected";
                        case 0:
                        case 1:
                            break;
                        }
                }
                return null;
            };

            /**
             * Creates a Hook message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.Hook
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.Hook} Hook
             */
            Hook.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.Hook)
                    return object;
                var message = new $root.hapi.release.Hook();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.kind != null)
                    message.kind = String(object.kind);
                if (object.path != null)
                    message.path = String(object.path);
                if (object.manifest != null)
                    message.manifest = String(object.manifest);
                if (object.events) {
                    if (!Array.isArray(object.events))
                        throw TypeError(".hapi.release.Hook.events: array expected");
                    message.events = [];
                    for (var i = 0; i < object.events.length; ++i)
                        switch (object.events[i]) {
                        default:
                        case "UNKNOWN":
                        case 0:
                            message.events[i] = 0;
                            break;
                        case "PRE_INSTALL":
                        case 1:
                            message.events[i] = 1;
                            break;
                        case "POST_INSTALL":
                        case 2:
                            message.events[i] = 2;
                            break;
                        case "PRE_DELETE":
                        case 3:
                            message.events[i] = 3;
                            break;
                        case "POST_DELETE":
                        case 4:
                            message.events[i] = 4;
                            break;
                        case "PRE_UPGRADE":
                        case 5:
                            message.events[i] = 5;
                            break;
                        case "POST_UPGRADE":
                        case 6:
                            message.events[i] = 6;
                            break;
                        case "PRE_ROLLBACK":
                        case 7:
                            message.events[i] = 7;
                            break;
                        case "POST_ROLLBACK":
                        case 8:
                            message.events[i] = 8;
                            break;
                        case "RELEASE_TEST_SUCCESS":
                        case 9:
                            message.events[i] = 9;
                            break;
                        case "RELEASE_TEST_FAILURE":
                        case 10:
                            message.events[i] = 10;
                            break;
                        }
                }
                if (object.lastRun != null) {
                    if (typeof object.lastRun !== "object")
                        throw TypeError(".hapi.release.Hook.lastRun: object expected");
                    message.lastRun = $root.google.protobuf.Timestamp.fromObject(object.lastRun);
                }
                if (object.weight != null)
                    message.weight = object.weight | 0;
                if (object.deletePolicies) {
                    if (!Array.isArray(object.deletePolicies))
                        throw TypeError(".hapi.release.Hook.deletePolicies: array expected");
                    message.deletePolicies = [];
                    for (var i = 0; i < object.deletePolicies.length; ++i)
                        switch (object.deletePolicies[i]) {
                        default:
                        case "SUCCEEDED":
                        case 0:
                            message.deletePolicies[i] = 0;
                            break;
                        case "FAILED":
                        case 1:
                            message.deletePolicies[i] = 1;
                            break;
                        }
                }
                return message;
            };

            /**
             * Creates a plain object from a Hook message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.Hook
             * @static
             * @param {hapi.release.Hook} message Hook
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Hook.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults) {
                    object.events = [];
                    object.deletePolicies = [];
                }
                if (options.defaults) {
                    object.name = "";
                    object.kind = "";
                    object.path = "";
                    object.manifest = "";
                    object.lastRun = null;
                    object.weight = 0;
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.kind != null && message.hasOwnProperty("kind"))
                    object.kind = message.kind;
                if (message.path != null && message.hasOwnProperty("path"))
                    object.path = message.path;
                if (message.manifest != null && message.hasOwnProperty("manifest"))
                    object.manifest = message.manifest;
                if (message.events && message.events.length) {
                    object.events = [];
                    for (var j = 0; j < message.events.length; ++j)
                        object.events[j] = options.enums === String ? $root.hapi.release.Hook.Event[message.events[j]] : message.events[j];
                }
                if (message.lastRun != null && message.hasOwnProperty("lastRun"))
                    object.lastRun = $root.google.protobuf.Timestamp.toObject(message.lastRun, options);
                if (message.weight != null && message.hasOwnProperty("weight"))
                    object.weight = message.weight;
                if (message.deletePolicies && message.deletePolicies.length) {
                    object.deletePolicies = [];
                    for (var j = 0; j < message.deletePolicies.length; ++j)
                        object.deletePolicies[j] = options.enums === String ? $root.hapi.release.Hook.DeletePolicy[message.deletePolicies[j]] : message.deletePolicies[j];
                }
                return object;
            };

            /**
             * Converts this Hook to JSON.
             * @function toJSON
             * @memberof hapi.release.Hook
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Hook.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Event enum.
             * @name hapi.release.Hook.Event
             * @enum {string}
             * @property {number} UNKNOWN=0 UNKNOWN value
             * @property {number} PRE_INSTALL=1 PRE_INSTALL value
             * @property {number} POST_INSTALL=2 POST_INSTALL value
             * @property {number} PRE_DELETE=3 PRE_DELETE value
             * @property {number} POST_DELETE=4 POST_DELETE value
             * @property {number} PRE_UPGRADE=5 PRE_UPGRADE value
             * @property {number} POST_UPGRADE=6 POST_UPGRADE value
             * @property {number} PRE_ROLLBACK=7 PRE_ROLLBACK value
             * @property {number} POST_ROLLBACK=8 POST_ROLLBACK value
             * @property {number} RELEASE_TEST_SUCCESS=9 RELEASE_TEST_SUCCESS value
             * @property {number} RELEASE_TEST_FAILURE=10 RELEASE_TEST_FAILURE value
             */
            Hook.Event = (function() {
                var valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "UNKNOWN"] = 0;
                values[valuesById[1] = "PRE_INSTALL"] = 1;
                values[valuesById[2] = "POST_INSTALL"] = 2;
                values[valuesById[3] = "PRE_DELETE"] = 3;
                values[valuesById[4] = "POST_DELETE"] = 4;
                values[valuesById[5] = "PRE_UPGRADE"] = 5;
                values[valuesById[6] = "POST_UPGRADE"] = 6;
                values[valuesById[7] = "PRE_ROLLBACK"] = 7;
                values[valuesById[8] = "POST_ROLLBACK"] = 8;
                values[valuesById[9] = "RELEASE_TEST_SUCCESS"] = 9;
                values[valuesById[10] = "RELEASE_TEST_FAILURE"] = 10;
                return values;
            })();

            /**
             * DeletePolicy enum.
             * @name hapi.release.Hook.DeletePolicy
             * @enum {string}
             * @property {number} SUCCEEDED=0 SUCCEEDED value
             * @property {number} FAILED=1 FAILED value
             */
            Hook.DeletePolicy = (function() {
                var valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "SUCCEEDED"] = 0;
                values[valuesById[1] = "FAILED"] = 1;
                return values;
            })();

            return Hook;
        })();

        release.Info = (function() {

            /**
             * Properties of an Info.
             * @memberof hapi.release
             * @interface IInfo
             * @property {hapi.release.IStatus|null} [status] Info status
             * @property {google.protobuf.ITimestamp|null} [firstDeployed] Info firstDeployed
             * @property {google.protobuf.ITimestamp|null} [lastDeployed] Info lastDeployed
             * @property {google.protobuf.ITimestamp|null} [deleted] Info deleted
             * @property {string|null} [Description] Info Description
             */

            /**
             * Constructs a new Info.
             * @memberof hapi.release
             * @classdesc Represents an Info.
             * @implements IInfo
             * @constructor
             * @param {hapi.release.IInfo=} [properties] Properties to set
             */
            function Info(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Info status.
             * @member {hapi.release.IStatus|null|undefined} status
             * @memberof hapi.release.Info
             * @instance
             */
            Info.prototype.status = null;

            /**
             * Info firstDeployed.
             * @member {google.protobuf.ITimestamp|null|undefined} firstDeployed
             * @memberof hapi.release.Info
             * @instance
             */
            Info.prototype.firstDeployed = null;

            /**
             * Info lastDeployed.
             * @member {google.protobuf.ITimestamp|null|undefined} lastDeployed
             * @memberof hapi.release.Info
             * @instance
             */
            Info.prototype.lastDeployed = null;

            /**
             * Info deleted.
             * @member {google.protobuf.ITimestamp|null|undefined} deleted
             * @memberof hapi.release.Info
             * @instance
             */
            Info.prototype.deleted = null;

            /**
             * Info Description.
             * @member {string} Description
             * @memberof hapi.release.Info
             * @instance
             */
            Info.prototype.Description = "";

            /**
             * Creates a new Info instance using the specified properties.
             * @function create
             * @memberof hapi.release.Info
             * @static
             * @param {hapi.release.IInfo=} [properties] Properties to set
             * @returns {hapi.release.Info} Info instance
             */
            Info.create = function create(properties) {
                return new Info(properties);
            };

            /**
             * Encodes the specified Info message. Does not implicitly {@link hapi.release.Info.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.Info
             * @static
             * @param {hapi.release.IInfo} message Info message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Info.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.status != null && message.hasOwnProperty("status"))
                    $root.hapi.release.Status.encode(message.status, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                if (message.firstDeployed != null && message.hasOwnProperty("firstDeployed"))
                    $root.google.protobuf.Timestamp.encode(message.firstDeployed, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.lastDeployed != null && message.hasOwnProperty("lastDeployed"))
                    $root.google.protobuf.Timestamp.encode(message.lastDeployed, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                if (message.deleted != null && message.hasOwnProperty("deleted"))
                    $root.google.protobuf.Timestamp.encode(message.deleted, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.Description != null && message.hasOwnProperty("Description"))
                    writer.uint32(/* id 5, wireType 2 =*/42).string(message.Description);
                return writer;
            };

            /**
             * Encodes the specified Info message, length delimited. Does not implicitly {@link hapi.release.Info.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.Info
             * @static
             * @param {hapi.release.IInfo} message Info message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Info.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes an Info message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.Info
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.Info} Info
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Info.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.Info();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.status = $root.hapi.release.Status.decode(reader, reader.uint32());
                        break;
                    case 2:
                        message.firstDeployed = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 3:
                        message.lastDeployed = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 4:
                        message.deleted = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 5:
                        message.Description = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes an Info message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.Info
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.Info} Info
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Info.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies an Info message.
             * @function verify
             * @memberof hapi.release.Info
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Info.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.status != null && message.hasOwnProperty("status")) {
                    var error = $root.hapi.release.Status.verify(message.status);
                    if (error)
                        return "status." + error;
                }
                if (message.firstDeployed != null && message.hasOwnProperty("firstDeployed")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.firstDeployed);
                    if (error)
                        return "firstDeployed." + error;
                }
                if (message.lastDeployed != null && message.hasOwnProperty("lastDeployed")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.lastDeployed);
                    if (error)
                        return "lastDeployed." + error;
                }
                if (message.deleted != null && message.hasOwnProperty("deleted")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.deleted);
                    if (error)
                        return "deleted." + error;
                }
                if (message.Description != null && message.hasOwnProperty("Description"))
                    if (!$util.isString(message.Description))
                        return "Description: string expected";
                return null;
            };

            /**
             * Creates an Info message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.Info
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.Info} Info
             */
            Info.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.Info)
                    return object;
                var message = new $root.hapi.release.Info();
                if (object.status != null) {
                    if (typeof object.status !== "object")
                        throw TypeError(".hapi.release.Info.status: object expected");
                    message.status = $root.hapi.release.Status.fromObject(object.status);
                }
                if (object.firstDeployed != null) {
                    if (typeof object.firstDeployed !== "object")
                        throw TypeError(".hapi.release.Info.firstDeployed: object expected");
                    message.firstDeployed = $root.google.protobuf.Timestamp.fromObject(object.firstDeployed);
                }
                if (object.lastDeployed != null) {
                    if (typeof object.lastDeployed !== "object")
                        throw TypeError(".hapi.release.Info.lastDeployed: object expected");
                    message.lastDeployed = $root.google.protobuf.Timestamp.fromObject(object.lastDeployed);
                }
                if (object.deleted != null) {
                    if (typeof object.deleted !== "object")
                        throw TypeError(".hapi.release.Info.deleted: object expected");
                    message.deleted = $root.google.protobuf.Timestamp.fromObject(object.deleted);
                }
                if (object.Description != null)
                    message.Description = String(object.Description);
                return message;
            };

            /**
             * Creates a plain object from an Info message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.Info
             * @static
             * @param {hapi.release.Info} message Info
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Info.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.status = null;
                    object.firstDeployed = null;
                    object.lastDeployed = null;
                    object.deleted = null;
                    object.Description = "";
                }
                if (message.status != null && message.hasOwnProperty("status"))
                    object.status = $root.hapi.release.Status.toObject(message.status, options);
                if (message.firstDeployed != null && message.hasOwnProperty("firstDeployed"))
                    object.firstDeployed = $root.google.protobuf.Timestamp.toObject(message.firstDeployed, options);
                if (message.lastDeployed != null && message.hasOwnProperty("lastDeployed"))
                    object.lastDeployed = $root.google.protobuf.Timestamp.toObject(message.lastDeployed, options);
                if (message.deleted != null && message.hasOwnProperty("deleted"))
                    object.deleted = $root.google.protobuf.Timestamp.toObject(message.deleted, options);
                if (message.Description != null && message.hasOwnProperty("Description"))
                    object.Description = message.Description;
                return object;
            };

            /**
             * Converts this Info to JSON.
             * @function toJSON
             * @memberof hapi.release.Info
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Info.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Info;
        })();

        release.Status = (function() {

            /**
             * Properties of a Status.
             * @memberof hapi.release
             * @interface IStatus
             * @property {hapi.release.Status.Code|null} [code] Status code
             * @property {string|null} [resources] Status resources
             * @property {string|null} [notes] Status notes
             * @property {hapi.release.ITestSuite|null} [lastTestSuiteRun] Status lastTestSuiteRun
             */

            /**
             * Constructs a new Status.
             * @memberof hapi.release
             * @classdesc Represents a Status.
             * @implements IStatus
             * @constructor
             * @param {hapi.release.IStatus=} [properties] Properties to set
             */
            function Status(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Status code.
             * @member {hapi.release.Status.Code} code
             * @memberof hapi.release.Status
             * @instance
             */
            Status.prototype.code = 0;

            /**
             * Status resources.
             * @member {string} resources
             * @memberof hapi.release.Status
             * @instance
             */
            Status.prototype.resources = "";

            /**
             * Status notes.
             * @member {string} notes
             * @memberof hapi.release.Status
             * @instance
             */
            Status.prototype.notes = "";

            /**
             * Status lastTestSuiteRun.
             * @member {hapi.release.ITestSuite|null|undefined} lastTestSuiteRun
             * @memberof hapi.release.Status
             * @instance
             */
            Status.prototype.lastTestSuiteRun = null;

            /**
             * Creates a new Status instance using the specified properties.
             * @function create
             * @memberof hapi.release.Status
             * @static
             * @param {hapi.release.IStatus=} [properties] Properties to set
             * @returns {hapi.release.Status} Status instance
             */
            Status.create = function create(properties) {
                return new Status(properties);
            };

            /**
             * Encodes the specified Status message. Does not implicitly {@link hapi.release.Status.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.Status
             * @static
             * @param {hapi.release.IStatus} message Status message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Status.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.code != null && message.hasOwnProperty("code"))
                    writer.uint32(/* id 1, wireType 0 =*/8).int32(message.code);
                if (message.resources != null && message.hasOwnProperty("resources"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.resources);
                if (message.notes != null && message.hasOwnProperty("notes"))
                    writer.uint32(/* id 4, wireType 2 =*/34).string(message.notes);
                if (message.lastTestSuiteRun != null && message.hasOwnProperty("lastTestSuiteRun"))
                    $root.hapi.release.TestSuite.encode(message.lastTestSuiteRun, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                return writer;
            };

            /**
             * Encodes the specified Status message, length delimited. Does not implicitly {@link hapi.release.Status.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.Status
             * @static
             * @param {hapi.release.IStatus} message Status message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Status.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Status message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.Status
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.Status} Status
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Status.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.Status();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.code = reader.int32();
                        break;
                    case 3:
                        message.resources = reader.string();
                        break;
                    case 4:
                        message.notes = reader.string();
                        break;
                    case 5:
                        message.lastTestSuiteRun = $root.hapi.release.TestSuite.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Status message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.Status
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.Status} Status
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Status.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Status message.
             * @function verify
             * @memberof hapi.release.Status
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Status.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.code != null && message.hasOwnProperty("code"))
                    switch (message.code) {
                    default:
                        return "code: enum value expected";
                    case 0:
                    case 1:
                    case 2:
                    case 3:
                    case 4:
                    case 5:
                    case 6:
                    case 7:
                    case 8:
                        break;
                    }
                if (message.resources != null && message.hasOwnProperty("resources"))
                    if (!$util.isString(message.resources))
                        return "resources: string expected";
                if (message.notes != null && message.hasOwnProperty("notes"))
                    if (!$util.isString(message.notes))
                        return "notes: string expected";
                if (message.lastTestSuiteRun != null && message.hasOwnProperty("lastTestSuiteRun")) {
                    var error = $root.hapi.release.TestSuite.verify(message.lastTestSuiteRun);
                    if (error)
                        return "lastTestSuiteRun." + error;
                }
                return null;
            };

            /**
             * Creates a Status message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.Status
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.Status} Status
             */
            Status.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.Status)
                    return object;
                var message = new $root.hapi.release.Status();
                switch (object.code) {
                case "UNKNOWN":
                case 0:
                    message.code = 0;
                    break;
                case "DEPLOYED":
                case 1:
                    message.code = 1;
                    break;
                case "DELETED":
                case 2:
                    message.code = 2;
                    break;
                case "SUPERSEDED":
                case 3:
                    message.code = 3;
                    break;
                case "FAILED":
                case 4:
                    message.code = 4;
                    break;
                case "DELETING":
                case 5:
                    message.code = 5;
                    break;
                case "PENDING_INSTALL":
                case 6:
                    message.code = 6;
                    break;
                case "PENDING_UPGRADE":
                case 7:
                    message.code = 7;
                    break;
                case "PENDING_ROLLBACK":
                case 8:
                    message.code = 8;
                    break;
                }
                if (object.resources != null)
                    message.resources = String(object.resources);
                if (object.notes != null)
                    message.notes = String(object.notes);
                if (object.lastTestSuiteRun != null) {
                    if (typeof object.lastTestSuiteRun !== "object")
                        throw TypeError(".hapi.release.Status.lastTestSuiteRun: object expected");
                    message.lastTestSuiteRun = $root.hapi.release.TestSuite.fromObject(object.lastTestSuiteRun);
                }
                return message;
            };

            /**
             * Creates a plain object from a Status message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.Status
             * @static
             * @param {hapi.release.Status} message Status
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Status.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.code = options.enums === String ? "UNKNOWN" : 0;
                    object.resources = "";
                    object.notes = "";
                    object.lastTestSuiteRun = null;
                }
                if (message.code != null && message.hasOwnProperty("code"))
                    object.code = options.enums === String ? $root.hapi.release.Status.Code[message.code] : message.code;
                if (message.resources != null && message.hasOwnProperty("resources"))
                    object.resources = message.resources;
                if (message.notes != null && message.hasOwnProperty("notes"))
                    object.notes = message.notes;
                if (message.lastTestSuiteRun != null && message.hasOwnProperty("lastTestSuiteRun"))
                    object.lastTestSuiteRun = $root.hapi.release.TestSuite.toObject(message.lastTestSuiteRun, options);
                return object;
            };

            /**
             * Converts this Status to JSON.
             * @function toJSON
             * @memberof hapi.release.Status
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Status.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Code enum.
             * @name hapi.release.Status.Code
             * @enum {string}
             * @property {number} UNKNOWN=0 UNKNOWN value
             * @property {number} DEPLOYED=1 DEPLOYED value
             * @property {number} DELETED=2 DELETED value
             * @property {number} SUPERSEDED=3 SUPERSEDED value
             * @property {number} FAILED=4 FAILED value
             * @property {number} DELETING=5 DELETING value
             * @property {number} PENDING_INSTALL=6 PENDING_INSTALL value
             * @property {number} PENDING_UPGRADE=7 PENDING_UPGRADE value
             * @property {number} PENDING_ROLLBACK=8 PENDING_ROLLBACK value
             */
            Status.Code = (function() {
                var valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "UNKNOWN"] = 0;
                values[valuesById[1] = "DEPLOYED"] = 1;
                values[valuesById[2] = "DELETED"] = 2;
                values[valuesById[3] = "SUPERSEDED"] = 3;
                values[valuesById[4] = "FAILED"] = 4;
                values[valuesById[5] = "DELETING"] = 5;
                values[valuesById[6] = "PENDING_INSTALL"] = 6;
                values[valuesById[7] = "PENDING_UPGRADE"] = 7;
                values[valuesById[8] = "PENDING_ROLLBACK"] = 8;
                return values;
            })();

            return Status;
        })();

        release.TestSuite = (function() {

            /**
             * Properties of a TestSuite.
             * @memberof hapi.release
             * @interface ITestSuite
             * @property {google.protobuf.ITimestamp|null} [startedAt] TestSuite startedAt
             * @property {google.protobuf.ITimestamp|null} [completedAt] TestSuite completedAt
             * @property {Array.<hapi.release.ITestRun>|null} [results] TestSuite results
             */

            /**
             * Constructs a new TestSuite.
             * @memberof hapi.release
             * @classdesc Represents a TestSuite.
             * @implements ITestSuite
             * @constructor
             * @param {hapi.release.ITestSuite=} [properties] Properties to set
             */
            function TestSuite(properties) {
                this.results = [];
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * TestSuite startedAt.
             * @member {google.protobuf.ITimestamp|null|undefined} startedAt
             * @memberof hapi.release.TestSuite
             * @instance
             */
            TestSuite.prototype.startedAt = null;

            /**
             * TestSuite completedAt.
             * @member {google.protobuf.ITimestamp|null|undefined} completedAt
             * @memberof hapi.release.TestSuite
             * @instance
             */
            TestSuite.prototype.completedAt = null;

            /**
             * TestSuite results.
             * @member {Array.<hapi.release.ITestRun>} results
             * @memberof hapi.release.TestSuite
             * @instance
             */
            TestSuite.prototype.results = $util.emptyArray;

            /**
             * Creates a new TestSuite instance using the specified properties.
             * @function create
             * @memberof hapi.release.TestSuite
             * @static
             * @param {hapi.release.ITestSuite=} [properties] Properties to set
             * @returns {hapi.release.TestSuite} TestSuite instance
             */
            TestSuite.create = function create(properties) {
                return new TestSuite(properties);
            };

            /**
             * Encodes the specified TestSuite message. Does not implicitly {@link hapi.release.TestSuite.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.TestSuite
             * @static
             * @param {hapi.release.ITestSuite} message TestSuite message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            TestSuite.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                    $root.google.protobuf.Timestamp.encode(message.startedAt, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                if (message.completedAt != null && message.hasOwnProperty("completedAt"))
                    $root.google.protobuf.Timestamp.encode(message.completedAt, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.results != null && message.results.length)
                    for (var i = 0; i < message.results.length; ++i)
                        $root.hapi.release.TestRun.encode(message.results[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                return writer;
            };

            /**
             * Encodes the specified TestSuite message, length delimited. Does not implicitly {@link hapi.release.TestSuite.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.TestSuite
             * @static
             * @param {hapi.release.ITestSuite} message TestSuite message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            TestSuite.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a TestSuite message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.TestSuite
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.TestSuite} TestSuite
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            TestSuite.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.TestSuite();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.startedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 2:
                        message.completedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 3:
                        if (!(message.results && message.results.length))
                            message.results = [];
                        message.results.push($root.hapi.release.TestRun.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a TestSuite message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.TestSuite
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.TestSuite} TestSuite
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            TestSuite.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a TestSuite message.
             * @function verify
             * @memberof hapi.release.TestSuite
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            TestSuite.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.startedAt != null && message.hasOwnProperty("startedAt")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.startedAt);
                    if (error)
                        return "startedAt." + error;
                }
                if (message.completedAt != null && message.hasOwnProperty("completedAt")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.completedAt);
                    if (error)
                        return "completedAt." + error;
                }
                if (message.results != null && message.hasOwnProperty("results")) {
                    if (!Array.isArray(message.results))
                        return "results: array expected";
                    for (var i = 0; i < message.results.length; ++i) {
                        var error = $root.hapi.release.TestRun.verify(message.results[i]);
                        if (error)
                            return "results." + error;
                    }
                }
                return null;
            };

            /**
             * Creates a TestSuite message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.TestSuite
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.TestSuite} TestSuite
             */
            TestSuite.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.TestSuite)
                    return object;
                var message = new $root.hapi.release.TestSuite();
                if (object.startedAt != null) {
                    if (typeof object.startedAt !== "object")
                        throw TypeError(".hapi.release.TestSuite.startedAt: object expected");
                    message.startedAt = $root.google.protobuf.Timestamp.fromObject(object.startedAt);
                }
                if (object.completedAt != null) {
                    if (typeof object.completedAt !== "object")
                        throw TypeError(".hapi.release.TestSuite.completedAt: object expected");
                    message.completedAt = $root.google.protobuf.Timestamp.fromObject(object.completedAt);
                }
                if (object.results) {
                    if (!Array.isArray(object.results))
                        throw TypeError(".hapi.release.TestSuite.results: array expected");
                    message.results = [];
                    for (var i = 0; i < object.results.length; ++i) {
                        if (typeof object.results[i] !== "object")
                            throw TypeError(".hapi.release.TestSuite.results: object expected");
                        message.results[i] = $root.hapi.release.TestRun.fromObject(object.results[i]);
                    }
                }
                return message;
            };

            /**
             * Creates a plain object from a TestSuite message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.TestSuite
             * @static
             * @param {hapi.release.TestSuite} message TestSuite
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            TestSuite.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults)
                    object.results = [];
                if (options.defaults) {
                    object.startedAt = null;
                    object.completedAt = null;
                }
                if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                    object.startedAt = $root.google.protobuf.Timestamp.toObject(message.startedAt, options);
                if (message.completedAt != null && message.hasOwnProperty("completedAt"))
                    object.completedAt = $root.google.protobuf.Timestamp.toObject(message.completedAt, options);
                if (message.results && message.results.length) {
                    object.results = [];
                    for (var j = 0; j < message.results.length; ++j)
                        object.results[j] = $root.hapi.release.TestRun.toObject(message.results[j], options);
                }
                return object;
            };

            /**
             * Converts this TestSuite to JSON.
             * @function toJSON
             * @memberof hapi.release.TestSuite
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            TestSuite.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return TestSuite;
        })();

        release.TestRun = (function() {

            /**
             * Properties of a TestRun.
             * @memberof hapi.release
             * @interface ITestRun
             * @property {string|null} [name] TestRun name
             * @property {hapi.release.TestRun.Status|null} [status] TestRun status
             * @property {string|null} [info] TestRun info
             * @property {google.protobuf.ITimestamp|null} [startedAt] TestRun startedAt
             * @property {google.protobuf.ITimestamp|null} [completedAt] TestRun completedAt
             */

            /**
             * Constructs a new TestRun.
             * @memberof hapi.release
             * @classdesc Represents a TestRun.
             * @implements ITestRun
             * @constructor
             * @param {hapi.release.ITestRun=} [properties] Properties to set
             */
            function TestRun(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * TestRun name.
             * @member {string} name
             * @memberof hapi.release.TestRun
             * @instance
             */
            TestRun.prototype.name = "";

            /**
             * TestRun status.
             * @member {hapi.release.TestRun.Status} status
             * @memberof hapi.release.TestRun
             * @instance
             */
            TestRun.prototype.status = 0;

            /**
             * TestRun info.
             * @member {string} info
             * @memberof hapi.release.TestRun
             * @instance
             */
            TestRun.prototype.info = "";

            /**
             * TestRun startedAt.
             * @member {google.protobuf.ITimestamp|null|undefined} startedAt
             * @memberof hapi.release.TestRun
             * @instance
             */
            TestRun.prototype.startedAt = null;

            /**
             * TestRun completedAt.
             * @member {google.protobuf.ITimestamp|null|undefined} completedAt
             * @memberof hapi.release.TestRun
             * @instance
             */
            TestRun.prototype.completedAt = null;

            /**
             * Creates a new TestRun instance using the specified properties.
             * @function create
             * @memberof hapi.release.TestRun
             * @static
             * @param {hapi.release.ITestRun=} [properties] Properties to set
             * @returns {hapi.release.TestRun} TestRun instance
             */
            TestRun.create = function create(properties) {
                return new TestRun(properties);
            };

            /**
             * Encodes the specified TestRun message. Does not implicitly {@link hapi.release.TestRun.verify|verify} messages.
             * @function encode
             * @memberof hapi.release.TestRun
             * @static
             * @param {hapi.release.ITestRun} message TestRun message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            TestRun.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.status != null && message.hasOwnProperty("status"))
                    writer.uint32(/* id 2, wireType 0 =*/16).int32(message.status);
                if (message.info != null && message.hasOwnProperty("info"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.info);
                if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                    $root.google.protobuf.Timestamp.encode(message.startedAt, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.completedAt != null && message.hasOwnProperty("completedAt"))
                    $root.google.protobuf.Timestamp.encode(message.completedAt, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                return writer;
            };

            /**
             * Encodes the specified TestRun message, length delimited. Does not implicitly {@link hapi.release.TestRun.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.release.TestRun
             * @static
             * @param {hapi.release.ITestRun} message TestRun message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            TestRun.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a TestRun message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.release.TestRun
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.release.TestRun} TestRun
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            TestRun.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.release.TestRun();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.status = reader.int32();
                        break;
                    case 3:
                        message.info = reader.string();
                        break;
                    case 4:
                        message.startedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    case 5:
                        message.completedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a TestRun message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.release.TestRun
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.release.TestRun} TestRun
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            TestRun.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a TestRun message.
             * @function verify
             * @memberof hapi.release.TestRun
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            TestRun.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.status != null && message.hasOwnProperty("status"))
                    switch (message.status) {
                    default:
                        return "status: enum value expected";
                    case 0:
                    case 1:
                    case 2:
                    case 3:
                        break;
                    }
                if (message.info != null && message.hasOwnProperty("info"))
                    if (!$util.isString(message.info))
                        return "info: string expected";
                if (message.startedAt != null && message.hasOwnProperty("startedAt")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.startedAt);
                    if (error)
                        return "startedAt." + error;
                }
                if (message.completedAt != null && message.hasOwnProperty("completedAt")) {
                    var error = $root.google.protobuf.Timestamp.verify(message.completedAt);
                    if (error)
                        return "completedAt." + error;
                }
                return null;
            };

            /**
             * Creates a TestRun message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.release.TestRun
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.release.TestRun} TestRun
             */
            TestRun.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.release.TestRun)
                    return object;
                var message = new $root.hapi.release.TestRun();
                if (object.name != null)
                    message.name = String(object.name);
                switch (object.status) {
                case "UNKNOWN":
                case 0:
                    message.status = 0;
                    break;
                case "SUCCESS":
                case 1:
                    message.status = 1;
                    break;
                case "FAILURE":
                case 2:
                    message.status = 2;
                    break;
                case "RUNNING":
                case 3:
                    message.status = 3;
                    break;
                }
                if (object.info != null)
                    message.info = String(object.info);
                if (object.startedAt != null) {
                    if (typeof object.startedAt !== "object")
                        throw TypeError(".hapi.release.TestRun.startedAt: object expected");
                    message.startedAt = $root.google.protobuf.Timestamp.fromObject(object.startedAt);
                }
                if (object.completedAt != null) {
                    if (typeof object.completedAt !== "object")
                        throw TypeError(".hapi.release.TestRun.completedAt: object expected");
                    message.completedAt = $root.google.protobuf.Timestamp.fromObject(object.completedAt);
                }
                return message;
            };

            /**
             * Creates a plain object from a TestRun message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.release.TestRun
             * @static
             * @param {hapi.release.TestRun} message TestRun
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            TestRun.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.name = "";
                    object.status = options.enums === String ? "UNKNOWN" : 0;
                    object.info = "";
                    object.startedAt = null;
                    object.completedAt = null;
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.status != null && message.hasOwnProperty("status"))
                    object.status = options.enums === String ? $root.hapi.release.TestRun.Status[message.status] : message.status;
                if (message.info != null && message.hasOwnProperty("info"))
                    object.info = message.info;
                if (message.startedAt != null && message.hasOwnProperty("startedAt"))
                    object.startedAt = $root.google.protobuf.Timestamp.toObject(message.startedAt, options);
                if (message.completedAt != null && message.hasOwnProperty("completedAt"))
                    object.completedAt = $root.google.protobuf.Timestamp.toObject(message.completedAt, options);
                return object;
            };

            /**
             * Converts this TestRun to JSON.
             * @function toJSON
             * @memberof hapi.release.TestRun
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            TestRun.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Status enum.
             * @name hapi.release.TestRun.Status
             * @enum {string}
             * @property {number} UNKNOWN=0 UNKNOWN value
             * @property {number} SUCCESS=1 SUCCESS value
             * @property {number} FAILURE=2 FAILURE value
             * @property {number} RUNNING=3 RUNNING value
             */
            TestRun.Status = (function() {
                var valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "UNKNOWN"] = 0;
                values[valuesById[1] = "SUCCESS"] = 1;
                values[valuesById[2] = "FAILURE"] = 2;
                values[valuesById[3] = "RUNNING"] = 3;
                return values;
            })();

            return TestRun;
        })();

        return release;
    })();

    hapi.chart = (function() {

        /**
         * Namespace chart.
         * @memberof hapi
         * @namespace
         */
        var chart = {};

        chart.Config = (function() {

            /**
             * Properties of a Config.
             * @memberof hapi.chart
             * @interface IConfig
             * @property {string|null} [raw] Config raw
             * @property {Object.<string,hapi.chart.IValue>|null} [values] Config values
             */

            /**
             * Constructs a new Config.
             * @memberof hapi.chart
             * @classdesc Represents a Config.
             * @implements IConfig
             * @constructor
             * @param {hapi.chart.IConfig=} [properties] Properties to set
             */
            function Config(properties) {
                this.values = {};
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Config raw.
             * @member {string} raw
             * @memberof hapi.chart.Config
             * @instance
             */
            Config.prototype.raw = "";

            /**
             * Config values.
             * @member {Object.<string,hapi.chart.IValue>} values
             * @memberof hapi.chart.Config
             * @instance
             */
            Config.prototype.values = $util.emptyObject;

            /**
             * Creates a new Config instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Config
             * @static
             * @param {hapi.chart.IConfig=} [properties] Properties to set
             * @returns {hapi.chart.Config} Config instance
             */
            Config.create = function create(properties) {
                return new Config(properties);
            };

            /**
             * Encodes the specified Config message. Does not implicitly {@link hapi.chart.Config.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Config
             * @static
             * @param {hapi.chart.IConfig} message Config message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Config.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.raw != null && message.hasOwnProperty("raw"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.raw);
                if (message.values != null && message.hasOwnProperty("values"))
                    for (var keys = Object.keys(message.values), i = 0; i < keys.length; ++i) {
                        writer.uint32(/* id 2, wireType 2 =*/18).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                        $root.hapi.chart.Value.encode(message.values[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                    }
                return writer;
            };

            /**
             * Encodes the specified Config message, length delimited. Does not implicitly {@link hapi.chart.Config.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Config
             * @static
             * @param {hapi.chart.IConfig} message Config message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Config.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Config message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Config
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Config} Config
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Config.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Config(), key;
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.raw = reader.string();
                        break;
                    case 2:
                        reader.skip().pos++;
                        if (message.values === $util.emptyObject)
                            message.values = {};
                        key = reader.string();
                        reader.pos++;
                        message.values[key] = $root.hapi.chart.Value.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Config message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Config
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Config} Config
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Config.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Config message.
             * @function verify
             * @memberof hapi.chart.Config
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Config.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.raw != null && message.hasOwnProperty("raw"))
                    if (!$util.isString(message.raw))
                        return "raw: string expected";
                if (message.values != null && message.hasOwnProperty("values")) {
                    if (!$util.isObject(message.values))
                        return "values: object expected";
                    var key = Object.keys(message.values);
                    for (var i = 0; i < key.length; ++i) {
                        var error = $root.hapi.chart.Value.verify(message.values[key[i]]);
                        if (error)
                            return "values." + error;
                    }
                }
                return null;
            };

            /**
             * Creates a Config message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Config
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Config} Config
             */
            Config.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Config)
                    return object;
                var message = new $root.hapi.chart.Config();
                if (object.raw != null)
                    message.raw = String(object.raw);
                if (object.values) {
                    if (typeof object.values !== "object")
                        throw TypeError(".hapi.chart.Config.values: object expected");
                    message.values = {};
                    for (var keys = Object.keys(object.values), i = 0; i < keys.length; ++i) {
                        if (typeof object.values[keys[i]] !== "object")
                            throw TypeError(".hapi.chart.Config.values: object expected");
                        message.values[keys[i]] = $root.hapi.chart.Value.fromObject(object.values[keys[i]]);
                    }
                }
                return message;
            };

            /**
             * Creates a plain object from a Config message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Config
             * @static
             * @param {hapi.chart.Config} message Config
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Config.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.objects || options.defaults)
                    object.values = {};
                if (options.defaults)
                    object.raw = "";
                if (message.raw != null && message.hasOwnProperty("raw"))
                    object.raw = message.raw;
                var keys2;
                if (message.values && (keys2 = Object.keys(message.values)).length) {
                    object.values = {};
                    for (var j = 0; j < keys2.length; ++j)
                        object.values[keys2[j]] = $root.hapi.chart.Value.toObject(message.values[keys2[j]], options);
                }
                return object;
            };

            /**
             * Converts this Config to JSON.
             * @function toJSON
             * @memberof hapi.chart.Config
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Config.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Config;
        })();

        chart.Value = (function() {

            /**
             * Properties of a Value.
             * @memberof hapi.chart
             * @interface IValue
             * @property {string|null} [value] Value value
             */

            /**
             * Constructs a new Value.
             * @memberof hapi.chart
             * @classdesc Represents a Value.
             * @implements IValue
             * @constructor
             * @param {hapi.chart.IValue=} [properties] Properties to set
             */
            function Value(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Value value.
             * @member {string} value
             * @memberof hapi.chart.Value
             * @instance
             */
            Value.prototype.value = "";

            /**
             * Creates a new Value instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Value
             * @static
             * @param {hapi.chart.IValue=} [properties] Properties to set
             * @returns {hapi.chart.Value} Value instance
             */
            Value.create = function create(properties) {
                return new Value(properties);
            };

            /**
             * Encodes the specified Value message. Does not implicitly {@link hapi.chart.Value.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Value
             * @static
             * @param {hapi.chart.IValue} message Value message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Value.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.value != null && message.hasOwnProperty("value"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.value);
                return writer;
            };

            /**
             * Encodes the specified Value message, length delimited. Does not implicitly {@link hapi.chart.Value.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Value
             * @static
             * @param {hapi.chart.IValue} message Value message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Value.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Value message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Value
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Value} Value
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Value.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Value();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.value = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Value message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Value
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Value} Value
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Value.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Value message.
             * @function verify
             * @memberof hapi.chart.Value
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Value.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.value != null && message.hasOwnProperty("value"))
                    if (!$util.isString(message.value))
                        return "value: string expected";
                return null;
            };

            /**
             * Creates a Value message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Value
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Value} Value
             */
            Value.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Value)
                    return object;
                var message = new $root.hapi.chart.Value();
                if (object.value != null)
                    message.value = String(object.value);
                return message;
            };

            /**
             * Creates a plain object from a Value message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Value
             * @static
             * @param {hapi.chart.Value} message Value
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Value.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults)
                    object.value = "";
                if (message.value != null && message.hasOwnProperty("value"))
                    object.value = message.value;
                return object;
            };

            /**
             * Converts this Value to JSON.
             * @function toJSON
             * @memberof hapi.chart.Value
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Value.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Value;
        })();

        chart.Chart = (function() {

            /**
             * Properties of a Chart.
             * @memberof hapi.chart
             * @interface IChart
             * @property {hapi.chart.IMetadata|null} [metadata] Chart metadata
             * @property {Array.<hapi.chart.ITemplate>|null} [templates] Chart templates
             * @property {Array.<hapi.chart.IChart>|null} [dependencies] Chart dependencies
             * @property {hapi.chart.IConfig|null} [values] Chart values
             * @property {Array.<google.protobuf.IAny>|null} [files] Chart files
             */

            /**
             * Constructs a new Chart.
             * @memberof hapi.chart
             * @classdesc Represents a Chart.
             * @implements IChart
             * @constructor
             * @param {hapi.chart.IChart=} [properties] Properties to set
             */
            function Chart(properties) {
                this.templates = [];
                this.dependencies = [];
                this.files = [];
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Chart metadata.
             * @member {hapi.chart.IMetadata|null|undefined} metadata
             * @memberof hapi.chart.Chart
             * @instance
             */
            Chart.prototype.metadata = null;

            /**
             * Chart templates.
             * @member {Array.<hapi.chart.ITemplate>} templates
             * @memberof hapi.chart.Chart
             * @instance
             */
            Chart.prototype.templates = $util.emptyArray;

            /**
             * Chart dependencies.
             * @member {Array.<hapi.chart.IChart>} dependencies
             * @memberof hapi.chart.Chart
             * @instance
             */
            Chart.prototype.dependencies = $util.emptyArray;

            /**
             * Chart values.
             * @member {hapi.chart.IConfig|null|undefined} values
             * @memberof hapi.chart.Chart
             * @instance
             */
            Chart.prototype.values = null;

            /**
             * Chart files.
             * @member {Array.<google.protobuf.IAny>} files
             * @memberof hapi.chart.Chart
             * @instance
             */
            Chart.prototype.files = $util.emptyArray;

            /**
             * Creates a new Chart instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Chart
             * @static
             * @param {hapi.chart.IChart=} [properties] Properties to set
             * @returns {hapi.chart.Chart} Chart instance
             */
            Chart.create = function create(properties) {
                return new Chart(properties);
            };

            /**
             * Encodes the specified Chart message. Does not implicitly {@link hapi.chart.Chart.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Chart
             * @static
             * @param {hapi.chart.IChart} message Chart message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Chart.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.metadata != null && message.hasOwnProperty("metadata"))
                    $root.hapi.chart.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                if (message.templates != null && message.templates.length)
                    for (var i = 0; i < message.templates.length; ++i)
                        $root.hapi.chart.Template.encode(message.templates[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.dependencies != null && message.dependencies.length)
                    for (var i = 0; i < message.dependencies.length; ++i)
                        $root.hapi.chart.Chart.encode(message.dependencies[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                if (message.values != null && message.hasOwnProperty("values"))
                    $root.hapi.chart.Config.encode(message.values, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.files != null && message.files.length)
                    for (var i = 0; i < message.files.length; ++i)
                        $root.google.protobuf.Any.encode(message.files[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                return writer;
            };

            /**
             * Encodes the specified Chart message, length delimited. Does not implicitly {@link hapi.chart.Chart.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Chart
             * @static
             * @param {hapi.chart.IChart} message Chart message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Chart.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Chart message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Chart
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Chart} Chart
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Chart.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Chart();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.metadata = $root.hapi.chart.Metadata.decode(reader, reader.uint32());
                        break;
                    case 2:
                        if (!(message.templates && message.templates.length))
                            message.templates = [];
                        message.templates.push($root.hapi.chart.Template.decode(reader, reader.uint32()));
                        break;
                    case 3:
                        if (!(message.dependencies && message.dependencies.length))
                            message.dependencies = [];
                        message.dependencies.push($root.hapi.chart.Chart.decode(reader, reader.uint32()));
                        break;
                    case 4:
                        message.values = $root.hapi.chart.Config.decode(reader, reader.uint32());
                        break;
                    case 5:
                        if (!(message.files && message.files.length))
                            message.files = [];
                        message.files.push($root.google.protobuf.Any.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Chart message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Chart
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Chart} Chart
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Chart.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Chart message.
             * @function verify
             * @memberof hapi.chart.Chart
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Chart.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.metadata != null && message.hasOwnProperty("metadata")) {
                    var error = $root.hapi.chart.Metadata.verify(message.metadata);
                    if (error)
                        return "metadata." + error;
                }
                if (message.templates != null && message.hasOwnProperty("templates")) {
                    if (!Array.isArray(message.templates))
                        return "templates: array expected";
                    for (var i = 0; i < message.templates.length; ++i) {
                        var error = $root.hapi.chart.Template.verify(message.templates[i]);
                        if (error)
                            return "templates." + error;
                    }
                }
                if (message.dependencies != null && message.hasOwnProperty("dependencies")) {
                    if (!Array.isArray(message.dependencies))
                        return "dependencies: array expected";
                    for (var i = 0; i < message.dependencies.length; ++i) {
                        var error = $root.hapi.chart.Chart.verify(message.dependencies[i]);
                        if (error)
                            return "dependencies." + error;
                    }
                }
                if (message.values != null && message.hasOwnProperty("values")) {
                    var error = $root.hapi.chart.Config.verify(message.values);
                    if (error)
                        return "values." + error;
                }
                if (message.files != null && message.hasOwnProperty("files")) {
                    if (!Array.isArray(message.files))
                        return "files: array expected";
                    for (var i = 0; i < message.files.length; ++i) {
                        var error = $root.google.protobuf.Any.verify(message.files[i]);
                        if (error)
                            return "files." + error;
                    }
                }
                return null;
            };

            /**
             * Creates a Chart message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Chart
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Chart} Chart
             */
            Chart.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Chart)
                    return object;
                var message = new $root.hapi.chart.Chart();
                if (object.metadata != null) {
                    if (typeof object.metadata !== "object")
                        throw TypeError(".hapi.chart.Chart.metadata: object expected");
                    message.metadata = $root.hapi.chart.Metadata.fromObject(object.metadata);
                }
                if (object.templates) {
                    if (!Array.isArray(object.templates))
                        throw TypeError(".hapi.chart.Chart.templates: array expected");
                    message.templates = [];
                    for (var i = 0; i < object.templates.length; ++i) {
                        if (typeof object.templates[i] !== "object")
                            throw TypeError(".hapi.chart.Chart.templates: object expected");
                        message.templates[i] = $root.hapi.chart.Template.fromObject(object.templates[i]);
                    }
                }
                if (object.dependencies) {
                    if (!Array.isArray(object.dependencies))
                        throw TypeError(".hapi.chart.Chart.dependencies: array expected");
                    message.dependencies = [];
                    for (var i = 0; i < object.dependencies.length; ++i) {
                        if (typeof object.dependencies[i] !== "object")
                            throw TypeError(".hapi.chart.Chart.dependencies: object expected");
                        message.dependencies[i] = $root.hapi.chart.Chart.fromObject(object.dependencies[i]);
                    }
                }
                if (object.values != null) {
                    if (typeof object.values !== "object")
                        throw TypeError(".hapi.chart.Chart.values: object expected");
                    message.values = $root.hapi.chart.Config.fromObject(object.values);
                }
                if (object.files) {
                    if (!Array.isArray(object.files))
                        throw TypeError(".hapi.chart.Chart.files: array expected");
                    message.files = [];
                    for (var i = 0; i < object.files.length; ++i) {
                        if (typeof object.files[i] !== "object")
                            throw TypeError(".hapi.chart.Chart.files: object expected");
                        message.files[i] = $root.google.protobuf.Any.fromObject(object.files[i]);
                    }
                }
                return message;
            };

            /**
             * Creates a plain object from a Chart message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Chart
             * @static
             * @param {hapi.chart.Chart} message Chart
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Chart.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults) {
                    object.templates = [];
                    object.dependencies = [];
                    object.files = [];
                }
                if (options.defaults) {
                    object.metadata = null;
                    object.values = null;
                }
                if (message.metadata != null && message.hasOwnProperty("metadata"))
                    object.metadata = $root.hapi.chart.Metadata.toObject(message.metadata, options);
                if (message.templates && message.templates.length) {
                    object.templates = [];
                    for (var j = 0; j < message.templates.length; ++j)
                        object.templates[j] = $root.hapi.chart.Template.toObject(message.templates[j], options);
                }
                if (message.dependencies && message.dependencies.length) {
                    object.dependencies = [];
                    for (var j = 0; j < message.dependencies.length; ++j)
                        object.dependencies[j] = $root.hapi.chart.Chart.toObject(message.dependencies[j], options);
                }
                if (message.values != null && message.hasOwnProperty("values"))
                    object.values = $root.hapi.chart.Config.toObject(message.values, options);
                if (message.files && message.files.length) {
                    object.files = [];
                    for (var j = 0; j < message.files.length; ++j)
                        object.files[j] = $root.google.protobuf.Any.toObject(message.files[j], options);
                }
                return object;
            };

            /**
             * Converts this Chart to JSON.
             * @function toJSON
             * @memberof hapi.chart.Chart
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Chart.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Chart;
        })();

        chart.Maintainer = (function() {

            /**
             * Properties of a Maintainer.
             * @memberof hapi.chart
             * @interface IMaintainer
             * @property {string|null} [name] Maintainer name
             * @property {string|null} [email] Maintainer email
             * @property {string|null} [url] Maintainer url
             */

            /**
             * Constructs a new Maintainer.
             * @memberof hapi.chart
             * @classdesc Represents a Maintainer.
             * @implements IMaintainer
             * @constructor
             * @param {hapi.chart.IMaintainer=} [properties] Properties to set
             */
            function Maintainer(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Maintainer name.
             * @member {string} name
             * @memberof hapi.chart.Maintainer
             * @instance
             */
            Maintainer.prototype.name = "";

            /**
             * Maintainer email.
             * @member {string} email
             * @memberof hapi.chart.Maintainer
             * @instance
             */
            Maintainer.prototype.email = "";

            /**
             * Maintainer url.
             * @member {string} url
             * @memberof hapi.chart.Maintainer
             * @instance
             */
            Maintainer.prototype.url = "";

            /**
             * Creates a new Maintainer instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {hapi.chart.IMaintainer=} [properties] Properties to set
             * @returns {hapi.chart.Maintainer} Maintainer instance
             */
            Maintainer.create = function create(properties) {
                return new Maintainer(properties);
            };

            /**
             * Encodes the specified Maintainer message. Does not implicitly {@link hapi.chart.Maintainer.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {hapi.chart.IMaintainer} message Maintainer message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Maintainer.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.email != null && message.hasOwnProperty("email"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.email);
                if (message.url != null && message.hasOwnProperty("url"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.url);
                return writer;
            };

            /**
             * Encodes the specified Maintainer message, length delimited. Does not implicitly {@link hapi.chart.Maintainer.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {hapi.chart.IMaintainer} message Maintainer message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Maintainer.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Maintainer message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Maintainer} Maintainer
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Maintainer.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Maintainer();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.email = reader.string();
                        break;
                    case 3:
                        message.url = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Maintainer message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Maintainer} Maintainer
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Maintainer.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Maintainer message.
             * @function verify
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Maintainer.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.email != null && message.hasOwnProperty("email"))
                    if (!$util.isString(message.email))
                        return "email: string expected";
                if (message.url != null && message.hasOwnProperty("url"))
                    if (!$util.isString(message.url))
                        return "url: string expected";
                return null;
            };

            /**
             * Creates a Maintainer message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Maintainer} Maintainer
             */
            Maintainer.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Maintainer)
                    return object;
                var message = new $root.hapi.chart.Maintainer();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.email != null)
                    message.email = String(object.email);
                if (object.url != null)
                    message.url = String(object.url);
                return message;
            };

            /**
             * Creates a plain object from a Maintainer message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Maintainer
             * @static
             * @param {hapi.chart.Maintainer} message Maintainer
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Maintainer.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.name = "";
                    object.email = "";
                    object.url = "";
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.email != null && message.hasOwnProperty("email"))
                    object.email = message.email;
                if (message.url != null && message.hasOwnProperty("url"))
                    object.url = message.url;
                return object;
            };

            /**
             * Converts this Maintainer to JSON.
             * @function toJSON
             * @memberof hapi.chart.Maintainer
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Maintainer.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Maintainer;
        })();

        chart.Metadata = (function() {

            /**
             * Properties of a Metadata.
             * @memberof hapi.chart
             * @interface IMetadata
             * @property {string|null} [name] Metadata name
             * @property {string|null} [home] Metadata home
             * @property {Array.<string>|null} [sources] Metadata sources
             * @property {string|null} [version] Metadata version
             * @property {string|null} [description] Metadata description
             * @property {Array.<string>|null} [keywords] Metadata keywords
             * @property {Array.<hapi.chart.IMaintainer>|null} [maintainers] Metadata maintainers
             * @property {string|null} [engine] Metadata engine
             * @property {string|null} [icon] Metadata icon
             * @property {string|null} [apiVersion] Metadata apiVersion
             * @property {string|null} [condition] Metadata condition
             * @property {string|null} [tags] Metadata tags
             * @property {string|null} [appVersion] Metadata appVersion
             * @property {boolean|null} [deprecated] Metadata deprecated
             * @property {string|null} [tillerVersion] Metadata tillerVersion
             * @property {Object.<string,string>|null} [annotations] Metadata annotations
             * @property {string|null} [kubeVersion] Metadata kubeVersion
             */

            /**
             * Constructs a new Metadata.
             * @memberof hapi.chart
             * @classdesc Represents a Metadata.
             * @implements IMetadata
             * @constructor
             * @param {hapi.chart.IMetadata=} [properties] Properties to set
             */
            function Metadata(properties) {
                this.sources = [];
                this.keywords = [];
                this.maintainers = [];
                this.annotations = {};
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Metadata name.
             * @member {string} name
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.name = "";

            /**
             * Metadata home.
             * @member {string} home
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.home = "";

            /**
             * Metadata sources.
             * @member {Array.<string>} sources
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.sources = $util.emptyArray;

            /**
             * Metadata version.
             * @member {string} version
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.version = "";

            /**
             * Metadata description.
             * @member {string} description
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.description = "";

            /**
             * Metadata keywords.
             * @member {Array.<string>} keywords
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.keywords = $util.emptyArray;

            /**
             * Metadata maintainers.
             * @member {Array.<hapi.chart.IMaintainer>} maintainers
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.maintainers = $util.emptyArray;

            /**
             * Metadata engine.
             * @member {string} engine
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.engine = "";

            /**
             * Metadata icon.
             * @member {string} icon
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.icon = "";

            /**
             * Metadata apiVersion.
             * @member {string} apiVersion
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.apiVersion = "";

            /**
             * Metadata condition.
             * @member {string} condition
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.condition = "";

            /**
             * Metadata tags.
             * @member {string} tags
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.tags = "";

            /**
             * Metadata appVersion.
             * @member {string} appVersion
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.appVersion = "";

            /**
             * Metadata deprecated.
             * @member {boolean} deprecated
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.deprecated = false;

            /**
             * Metadata tillerVersion.
             * @member {string} tillerVersion
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.tillerVersion = "";

            /**
             * Metadata annotations.
             * @member {Object.<string,string>} annotations
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.annotations = $util.emptyObject;

            /**
             * Metadata kubeVersion.
             * @member {string} kubeVersion
             * @memberof hapi.chart.Metadata
             * @instance
             */
            Metadata.prototype.kubeVersion = "";

            /**
             * Creates a new Metadata instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Metadata
             * @static
             * @param {hapi.chart.IMetadata=} [properties] Properties to set
             * @returns {hapi.chart.Metadata} Metadata instance
             */
            Metadata.create = function create(properties) {
                return new Metadata(properties);
            };

            /**
             * Encodes the specified Metadata message. Does not implicitly {@link hapi.chart.Metadata.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Metadata
             * @static
             * @param {hapi.chart.IMetadata} message Metadata message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Metadata.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.home != null && message.hasOwnProperty("home"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.home);
                if (message.sources != null && message.sources.length)
                    for (var i = 0; i < message.sources.length; ++i)
                        writer.uint32(/* id 3, wireType 2 =*/26).string(message.sources[i]);
                if (message.version != null && message.hasOwnProperty("version"))
                    writer.uint32(/* id 4, wireType 2 =*/34).string(message.version);
                if (message.description != null && message.hasOwnProperty("description"))
                    writer.uint32(/* id 5, wireType 2 =*/42).string(message.description);
                if (message.keywords != null && message.keywords.length)
                    for (var i = 0; i < message.keywords.length; ++i)
                        writer.uint32(/* id 6, wireType 2 =*/50).string(message.keywords[i]);
                if (message.maintainers != null && message.maintainers.length)
                    for (var i = 0; i < message.maintainers.length; ++i)
                        $root.hapi.chart.Maintainer.encode(message.maintainers[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
                if (message.engine != null && message.hasOwnProperty("engine"))
                    writer.uint32(/* id 8, wireType 2 =*/66).string(message.engine);
                if (message.icon != null && message.hasOwnProperty("icon"))
                    writer.uint32(/* id 9, wireType 2 =*/74).string(message.icon);
                if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                    writer.uint32(/* id 10, wireType 2 =*/82).string(message.apiVersion);
                if (message.condition != null && message.hasOwnProperty("condition"))
                    writer.uint32(/* id 11, wireType 2 =*/90).string(message.condition);
                if (message.tags != null && message.hasOwnProperty("tags"))
                    writer.uint32(/* id 12, wireType 2 =*/98).string(message.tags);
                if (message.appVersion != null && message.hasOwnProperty("appVersion"))
                    writer.uint32(/* id 13, wireType 2 =*/106).string(message.appVersion);
                if (message.deprecated != null && message.hasOwnProperty("deprecated"))
                    writer.uint32(/* id 14, wireType 0 =*/112).bool(message.deprecated);
                if (message.tillerVersion != null && message.hasOwnProperty("tillerVersion"))
                    writer.uint32(/* id 15, wireType 2 =*/122).string(message.tillerVersion);
                if (message.annotations != null && message.hasOwnProperty("annotations"))
                    for (var keys = Object.keys(message.annotations), i = 0; i < keys.length; ++i)
                        writer.uint32(/* id 16, wireType 2 =*/130).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 2 =*/18).string(message.annotations[keys[i]]).ldelim();
                if (message.kubeVersion != null && message.hasOwnProperty("kubeVersion"))
                    writer.uint32(/* id 17, wireType 2 =*/138).string(message.kubeVersion);
                return writer;
            };

            /**
             * Encodes the specified Metadata message, length delimited. Does not implicitly {@link hapi.chart.Metadata.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Metadata
             * @static
             * @param {hapi.chart.IMetadata} message Metadata message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Metadata.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Metadata message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Metadata
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Metadata} Metadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Metadata.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Metadata(), key;
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.home = reader.string();
                        break;
                    case 3:
                        if (!(message.sources && message.sources.length))
                            message.sources = [];
                        message.sources.push(reader.string());
                        break;
                    case 4:
                        message.version = reader.string();
                        break;
                    case 5:
                        message.description = reader.string();
                        break;
                    case 6:
                        if (!(message.keywords && message.keywords.length))
                            message.keywords = [];
                        message.keywords.push(reader.string());
                        break;
                    case 7:
                        if (!(message.maintainers && message.maintainers.length))
                            message.maintainers = [];
                        message.maintainers.push($root.hapi.chart.Maintainer.decode(reader, reader.uint32()));
                        break;
                    case 8:
                        message.engine = reader.string();
                        break;
                    case 9:
                        message.icon = reader.string();
                        break;
                    case 10:
                        message.apiVersion = reader.string();
                        break;
                    case 11:
                        message.condition = reader.string();
                        break;
                    case 12:
                        message.tags = reader.string();
                        break;
                    case 13:
                        message.appVersion = reader.string();
                        break;
                    case 14:
                        message.deprecated = reader.bool();
                        break;
                    case 15:
                        message.tillerVersion = reader.string();
                        break;
                    case 16:
                        reader.skip().pos++;
                        if (message.annotations === $util.emptyObject)
                            message.annotations = {};
                        key = reader.string();
                        reader.pos++;
                        message.annotations[key] = reader.string();
                        break;
                    case 17:
                        message.kubeVersion = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Metadata message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Metadata
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Metadata} Metadata
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Metadata.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Metadata message.
             * @function verify
             * @memberof hapi.chart.Metadata
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Metadata.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.home != null && message.hasOwnProperty("home"))
                    if (!$util.isString(message.home))
                        return "home: string expected";
                if (message.sources != null && message.hasOwnProperty("sources")) {
                    if (!Array.isArray(message.sources))
                        return "sources: array expected";
                    for (var i = 0; i < message.sources.length; ++i)
                        if (!$util.isString(message.sources[i]))
                            return "sources: string[] expected";
                }
                if (message.version != null && message.hasOwnProperty("version"))
                    if (!$util.isString(message.version))
                        return "version: string expected";
                if (message.description != null && message.hasOwnProperty("description"))
                    if (!$util.isString(message.description))
                        return "description: string expected";
                if (message.keywords != null && message.hasOwnProperty("keywords")) {
                    if (!Array.isArray(message.keywords))
                        return "keywords: array expected";
                    for (var i = 0; i < message.keywords.length; ++i)
                        if (!$util.isString(message.keywords[i]))
                            return "keywords: string[] expected";
                }
                if (message.maintainers != null && message.hasOwnProperty("maintainers")) {
                    if (!Array.isArray(message.maintainers))
                        return "maintainers: array expected";
                    for (var i = 0; i < message.maintainers.length; ++i) {
                        var error = $root.hapi.chart.Maintainer.verify(message.maintainers[i]);
                        if (error)
                            return "maintainers." + error;
                    }
                }
                if (message.engine != null && message.hasOwnProperty("engine"))
                    if (!$util.isString(message.engine))
                        return "engine: string expected";
                if (message.icon != null && message.hasOwnProperty("icon"))
                    if (!$util.isString(message.icon))
                        return "icon: string expected";
                if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                    if (!$util.isString(message.apiVersion))
                        return "apiVersion: string expected";
                if (message.condition != null && message.hasOwnProperty("condition"))
                    if (!$util.isString(message.condition))
                        return "condition: string expected";
                if (message.tags != null && message.hasOwnProperty("tags"))
                    if (!$util.isString(message.tags))
                        return "tags: string expected";
                if (message.appVersion != null && message.hasOwnProperty("appVersion"))
                    if (!$util.isString(message.appVersion))
                        return "appVersion: string expected";
                if (message.deprecated != null && message.hasOwnProperty("deprecated"))
                    if (typeof message.deprecated !== "boolean")
                        return "deprecated: boolean expected";
                if (message.tillerVersion != null && message.hasOwnProperty("tillerVersion"))
                    if (!$util.isString(message.tillerVersion))
                        return "tillerVersion: string expected";
                if (message.annotations != null && message.hasOwnProperty("annotations")) {
                    if (!$util.isObject(message.annotations))
                        return "annotations: object expected";
                    var key = Object.keys(message.annotations);
                    for (var i = 0; i < key.length; ++i)
                        if (!$util.isString(message.annotations[key[i]]))
                            return "annotations: string{k:string} expected";
                }
                if (message.kubeVersion != null && message.hasOwnProperty("kubeVersion"))
                    if (!$util.isString(message.kubeVersion))
                        return "kubeVersion: string expected";
                return null;
            };

            /**
             * Creates a Metadata message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Metadata
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Metadata} Metadata
             */
            Metadata.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Metadata)
                    return object;
                var message = new $root.hapi.chart.Metadata();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.home != null)
                    message.home = String(object.home);
                if (object.sources) {
                    if (!Array.isArray(object.sources))
                        throw TypeError(".hapi.chart.Metadata.sources: array expected");
                    message.sources = [];
                    for (var i = 0; i < object.sources.length; ++i)
                        message.sources[i] = String(object.sources[i]);
                }
                if (object.version != null)
                    message.version = String(object.version);
                if (object.description != null)
                    message.description = String(object.description);
                if (object.keywords) {
                    if (!Array.isArray(object.keywords))
                        throw TypeError(".hapi.chart.Metadata.keywords: array expected");
                    message.keywords = [];
                    for (var i = 0; i < object.keywords.length; ++i)
                        message.keywords[i] = String(object.keywords[i]);
                }
                if (object.maintainers) {
                    if (!Array.isArray(object.maintainers))
                        throw TypeError(".hapi.chart.Metadata.maintainers: array expected");
                    message.maintainers = [];
                    for (var i = 0; i < object.maintainers.length; ++i) {
                        if (typeof object.maintainers[i] !== "object")
                            throw TypeError(".hapi.chart.Metadata.maintainers: object expected");
                        message.maintainers[i] = $root.hapi.chart.Maintainer.fromObject(object.maintainers[i]);
                    }
                }
                if (object.engine != null)
                    message.engine = String(object.engine);
                if (object.icon != null)
                    message.icon = String(object.icon);
                if (object.apiVersion != null)
                    message.apiVersion = String(object.apiVersion);
                if (object.condition != null)
                    message.condition = String(object.condition);
                if (object.tags != null)
                    message.tags = String(object.tags);
                if (object.appVersion != null)
                    message.appVersion = String(object.appVersion);
                if (object.deprecated != null)
                    message.deprecated = Boolean(object.deprecated);
                if (object.tillerVersion != null)
                    message.tillerVersion = String(object.tillerVersion);
                if (object.annotations) {
                    if (typeof object.annotations !== "object")
                        throw TypeError(".hapi.chart.Metadata.annotations: object expected");
                    message.annotations = {};
                    for (var keys = Object.keys(object.annotations), i = 0; i < keys.length; ++i)
                        message.annotations[keys[i]] = String(object.annotations[keys[i]]);
                }
                if (object.kubeVersion != null)
                    message.kubeVersion = String(object.kubeVersion);
                return message;
            };

            /**
             * Creates a plain object from a Metadata message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Metadata
             * @static
             * @param {hapi.chart.Metadata} message Metadata
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Metadata.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults) {
                    object.sources = [];
                    object.keywords = [];
                    object.maintainers = [];
                }
                if (options.objects || options.defaults)
                    object.annotations = {};
                if (options.defaults) {
                    object.name = "";
                    object.home = "";
                    object.version = "";
                    object.description = "";
                    object.engine = "";
                    object.icon = "";
                    object.apiVersion = "";
                    object.condition = "";
                    object.tags = "";
                    object.appVersion = "";
                    object.deprecated = false;
                    object.tillerVersion = "";
                    object.kubeVersion = "";
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.home != null && message.hasOwnProperty("home"))
                    object.home = message.home;
                if (message.sources && message.sources.length) {
                    object.sources = [];
                    for (var j = 0; j < message.sources.length; ++j)
                        object.sources[j] = message.sources[j];
                }
                if (message.version != null && message.hasOwnProperty("version"))
                    object.version = message.version;
                if (message.description != null && message.hasOwnProperty("description"))
                    object.description = message.description;
                if (message.keywords && message.keywords.length) {
                    object.keywords = [];
                    for (var j = 0; j < message.keywords.length; ++j)
                        object.keywords[j] = message.keywords[j];
                }
                if (message.maintainers && message.maintainers.length) {
                    object.maintainers = [];
                    for (var j = 0; j < message.maintainers.length; ++j)
                        object.maintainers[j] = $root.hapi.chart.Maintainer.toObject(message.maintainers[j], options);
                }
                if (message.engine != null && message.hasOwnProperty("engine"))
                    object.engine = message.engine;
                if (message.icon != null && message.hasOwnProperty("icon"))
                    object.icon = message.icon;
                if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                    object.apiVersion = message.apiVersion;
                if (message.condition != null && message.hasOwnProperty("condition"))
                    object.condition = message.condition;
                if (message.tags != null && message.hasOwnProperty("tags"))
                    object.tags = message.tags;
                if (message.appVersion != null && message.hasOwnProperty("appVersion"))
                    object.appVersion = message.appVersion;
                if (message.deprecated != null && message.hasOwnProperty("deprecated"))
                    object.deprecated = message.deprecated;
                if (message.tillerVersion != null && message.hasOwnProperty("tillerVersion"))
                    object.tillerVersion = message.tillerVersion;
                var keys2;
                if (message.annotations && (keys2 = Object.keys(message.annotations)).length) {
                    object.annotations = {};
                    for (var j = 0; j < keys2.length; ++j)
                        object.annotations[keys2[j]] = message.annotations[keys2[j]];
                }
                if (message.kubeVersion != null && message.hasOwnProperty("kubeVersion"))
                    object.kubeVersion = message.kubeVersion;
                return object;
            };

            /**
             * Converts this Metadata to JSON.
             * @function toJSON
             * @memberof hapi.chart.Metadata
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Metadata.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Engine enum.
             * @name hapi.chart.Metadata.Engine
             * @enum {string}
             * @property {number} UNKNOWN=0 UNKNOWN value
             * @property {number} GOTPL=1 GOTPL value
             */
            Metadata.Engine = (function() {
                var valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "UNKNOWN"] = 0;
                values[valuesById[1] = "GOTPL"] = 1;
                return values;
            })();

            return Metadata;
        })();

        chart.Template = (function() {

            /**
             * Properties of a Template.
             * @memberof hapi.chart
             * @interface ITemplate
             * @property {string|null} [name] Template name
             * @property {Uint8Array|null} [data] Template data
             */

            /**
             * Constructs a new Template.
             * @memberof hapi.chart
             * @classdesc Represents a Template.
             * @implements ITemplate
             * @constructor
             * @param {hapi.chart.ITemplate=} [properties] Properties to set
             */
            function Template(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Template name.
             * @member {string} name
             * @memberof hapi.chart.Template
             * @instance
             */
            Template.prototype.name = "";

            /**
             * Template data.
             * @member {Uint8Array} data
             * @memberof hapi.chart.Template
             * @instance
             */
            Template.prototype.data = $util.newBuffer([]);

            /**
             * Creates a new Template instance using the specified properties.
             * @function create
             * @memberof hapi.chart.Template
             * @static
             * @param {hapi.chart.ITemplate=} [properties] Properties to set
             * @returns {hapi.chart.Template} Template instance
             */
            Template.create = function create(properties) {
                return new Template(properties);
            };

            /**
             * Encodes the specified Template message. Does not implicitly {@link hapi.chart.Template.verify|verify} messages.
             * @function encode
             * @memberof hapi.chart.Template
             * @static
             * @param {hapi.chart.ITemplate} message Template message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Template.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.hasOwnProperty("name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.data != null && message.hasOwnProperty("data"))
                    writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.data);
                return writer;
            };

            /**
             * Encodes the specified Template message, length delimited. Does not implicitly {@link hapi.chart.Template.verify|verify} messages.
             * @function encodeDelimited
             * @memberof hapi.chart.Template
             * @static
             * @param {hapi.chart.ITemplate} message Template message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Template.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Template message from the specified reader or buffer.
             * @function decode
             * @memberof hapi.chart.Template
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {hapi.chart.Template} Template
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Template.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.hapi.chart.Template();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.data = reader.bytes();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Template message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof hapi.chart.Template
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {hapi.chart.Template} Template
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Template.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Template message.
             * @function verify
             * @memberof hapi.chart.Template
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Template.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.name != null && message.hasOwnProperty("name"))
                    if (!$util.isString(message.name))
                        return "name: string expected";
                if (message.data != null && message.hasOwnProperty("data"))
                    if (!(message.data && typeof message.data.length === "number" || $util.isString(message.data)))
                        return "data: buffer expected";
                return null;
            };

            /**
             * Creates a Template message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof hapi.chart.Template
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {hapi.chart.Template} Template
             */
            Template.fromObject = function fromObject(object) {
                if (object instanceof $root.hapi.chart.Template)
                    return object;
                var message = new $root.hapi.chart.Template();
                if (object.name != null)
                    message.name = String(object.name);
                if (object.data != null)
                    if (typeof object.data === "string")
                        $util.base64.decode(object.data, message.data = $util.newBuffer($util.base64.length(object.data)), 0);
                    else if (object.data.length)
                        message.data = object.data;
                return message;
            };

            /**
             * Creates a plain object from a Template message. Also converts values to other types if specified.
             * @function toObject
             * @memberof hapi.chart.Template
             * @static
             * @param {hapi.chart.Template} message Template
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Template.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.name = "";
                    object.data = options.bytes === String ? "" : [];
                }
                if (message.name != null && message.hasOwnProperty("name"))
                    object.name = message.name;
                if (message.data != null && message.hasOwnProperty("data"))
                    object.data = options.bytes === String ? $util.base64.encode(message.data, 0, message.data.length) : options.bytes === Array ? Array.prototype.slice.call(message.data) : message.data;
                return object;
            };

            /**
             * Converts this Template to JSON.
             * @function toJSON
             * @memberof hapi.chart.Template
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Template.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Template;
        })();

        return chart;
    })();

    return hapi;
})();

$root.google = (function() {

    /**
     * Namespace google.
     * @exports google
     * @namespace
     */
    var google = {};

    google.protobuf = (function() {

        /**
         * Namespace protobuf.
         * @memberof google
         * @namespace
         */
        var protobuf = {};

        protobuf.Timestamp = (function() {

            /**
             * Properties of a Timestamp.
             * @memberof google.protobuf
             * @interface ITimestamp
             * @property {number|Long|null} [seconds] Timestamp seconds
             * @property {number|null} [nanos] Timestamp nanos
             */

            /**
             * Constructs a new Timestamp.
             * @memberof google.protobuf
             * @classdesc Represents a Timestamp.
             * @implements ITimestamp
             * @constructor
             * @param {google.protobuf.ITimestamp=} [properties] Properties to set
             */
            function Timestamp(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Timestamp seconds.
             * @member {number|Long} seconds
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

            /**
             * Timestamp nanos.
             * @member {number} nanos
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.nanos = 0;

            /**
             * Creates a new Timestamp instance using the specified properties.
             * @function create
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp=} [properties] Properties to set
             * @returns {google.protobuf.Timestamp} Timestamp instance
             */
            Timestamp.create = function create(properties) {
                return new Timestamp(properties);
            };

            /**
             * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Timestamp.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.seconds != null && message.hasOwnProperty("seconds"))
                    writer.uint32(/* id 1, wireType 0 =*/8).int64(message.seconds);
                if (message.nanos != null && message.hasOwnProperty("nanos"))
                    writer.uint32(/* id 2, wireType 0 =*/16).int32(message.nanos);
                return writer;
            };

            /**
             * Encodes the specified Timestamp message, length delimited. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @function encodeDelimited
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Timestamp.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes a Timestamp message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Timestamp} Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Timestamp.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Timestamp();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.seconds = reader.int64();
                        break;
                    case 2:
                        message.nanos = reader.int32();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes a Timestamp message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {google.protobuf.Timestamp} Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Timestamp.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies a Timestamp message.
             * @function verify
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Timestamp.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.seconds != null && message.hasOwnProperty("seconds"))
                    if (!$util.isInteger(message.seconds) && !(message.seconds && $util.isInteger(message.seconds.low) && $util.isInteger(message.seconds.high)))
                        return "seconds: integer|Long expected";
                if (message.nanos != null && message.hasOwnProperty("nanos"))
                    if (!$util.isInteger(message.nanos))
                        return "nanos: integer expected";
                return null;
            };

            /**
             * Creates a Timestamp message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {google.protobuf.Timestamp} Timestamp
             */
            Timestamp.fromObject = function fromObject(object) {
                if (object instanceof $root.google.protobuf.Timestamp)
                    return object;
                var message = new $root.google.protobuf.Timestamp();
                if (object.seconds != null)
                    if ($util.Long)
                        (message.seconds = $util.Long.fromValue(object.seconds)).unsigned = false;
                    else if (typeof object.seconds === "string")
                        message.seconds = parseInt(object.seconds, 10);
                    else if (typeof object.seconds === "number")
                        message.seconds = object.seconds;
                    else if (typeof object.seconds === "object")
                        message.seconds = new $util.LongBits(object.seconds.low >>> 0, object.seconds.high >>> 0).toNumber();
                if (object.nanos != null)
                    message.nanos = object.nanos | 0;
                return message;
            };

            /**
             * Creates a plain object from a Timestamp message. Also converts values to other types if specified.
             * @function toObject
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.Timestamp} message Timestamp
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Timestamp.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    if ($util.Long) {
                        var long = new $util.Long(0, 0, false);
                        object.seconds = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                    } else
                        object.seconds = options.longs === String ? "0" : 0;
                    object.nanos = 0;
                }
                if (message.seconds != null && message.hasOwnProperty("seconds"))
                    if (typeof message.seconds === "number")
                        object.seconds = options.longs === String ? String(message.seconds) : message.seconds;
                    else
                        object.seconds = options.longs === String ? $util.Long.prototype.toString.call(message.seconds) : options.longs === Number ? new $util.LongBits(message.seconds.low >>> 0, message.seconds.high >>> 0).toNumber() : message.seconds;
                if (message.nanos != null && message.hasOwnProperty("nanos"))
                    object.nanos = message.nanos;
                return object;
            };

            /**
             * Converts this Timestamp to JSON.
             * @function toJSON
             * @memberof google.protobuf.Timestamp
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Timestamp.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Timestamp;
        })();

        protobuf.Any = (function() {

            /**
             * Properties of an Any.
             * @memberof google.protobuf
             * @interface IAny
             * @property {string|null} [type_url] Any type_url
             * @property {Uint8Array|null} [value] Any value
             */

            /**
             * Constructs a new Any.
             * @memberof google.protobuf
             * @classdesc Represents an Any.
             * @implements IAny
             * @constructor
             * @param {google.protobuf.IAny=} [properties] Properties to set
             */
            function Any(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Any type_url.
             * @member {string} type_url
             * @memberof google.protobuf.Any
             * @instance
             */
            Any.prototype.type_url = "";

            /**
             * Any value.
             * @member {Uint8Array} value
             * @memberof google.protobuf.Any
             * @instance
             */
            Any.prototype.value = $util.newBuffer([]);

            /**
             * Creates a new Any instance using the specified properties.
             * @function create
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny=} [properties] Properties to set
             * @returns {google.protobuf.Any} Any instance
             */
            Any.create = function create(properties) {
                return new Any(properties);
            };

            /**
             * Encodes the specified Any message. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny} message Any message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Any.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.type_url != null && message.hasOwnProperty("type_url"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.type_url);
                if (message.value != null && message.hasOwnProperty("value"))
                    writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.value);
                return writer;
            };

            /**
             * Encodes the specified Any message, length delimited. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
             * @function encodeDelimited
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.IAny} message Any message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Any.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes an Any message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Any
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Any} Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Any.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Any();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.type_url = reader.string();
                        break;
                    case 2:
                        message.value = reader.bytes();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes an Any message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof google.protobuf.Any
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {google.protobuf.Any} Any
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Any.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies an Any message.
             * @function verify
             * @memberof google.protobuf.Any
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Any.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.type_url != null && message.hasOwnProperty("type_url"))
                    if (!$util.isString(message.type_url))
                        return "type_url: string expected";
                if (message.value != null && message.hasOwnProperty("value"))
                    if (!(message.value && typeof message.value.length === "number" || $util.isString(message.value)))
                        return "value: buffer expected";
                return null;
            };

            /**
             * Creates an Any message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof google.protobuf.Any
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {google.protobuf.Any} Any
             */
            Any.fromObject = function fromObject(object) {
                if (object instanceof $root.google.protobuf.Any)
                    return object;
                var message = new $root.google.protobuf.Any();
                if (object.type_url != null)
                    message.type_url = String(object.type_url);
                if (object.value != null)
                    if (typeof object.value === "string")
                        $util.base64.decode(object.value, message.value = $util.newBuffer($util.base64.length(object.value)), 0);
                    else if (object.value.length)
                        message.value = object.value;
                return message;
            };

            /**
             * Creates a plain object from an Any message. Also converts values to other types if specified.
             * @function toObject
             * @memberof google.protobuf.Any
             * @static
             * @param {google.protobuf.Any} message Any
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Any.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.type_url = "";
                    object.value = options.bytes === String ? "" : [];
                }
                if (message.type_url != null && message.hasOwnProperty("type_url"))
                    object.type_url = message.type_url;
                if (message.value != null && message.hasOwnProperty("value"))
                    object.value = options.bytes === String ? $util.base64.encode(message.value, 0, message.value.length) : options.bytes === Array ? Array.prototype.slice.call(message.value) : message.value;
                return object;
            };

            /**
             * Converts this Any to JSON.
             * @function toJSON
             * @memberof google.protobuf.Any
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Any.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return Any;
        })();

        return protobuf;
    })();

    return google;
})();

module.exports = $root;

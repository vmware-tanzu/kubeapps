/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "google.protobuf";

/**
 * The protocol compiler can output a FileDescriptorSet containing the .proto
 * files it parses.
 */
export interface FileDescriptorSet {
  file: FileDescriptorProto[];
}

/** Describes a complete .proto file. */
export interface FileDescriptorProto {
  /** file name, relative to root of source tree */
  name: string;
  /** e.g. "foo", "foo.bar", etc. */
  package: string;
  /** Names of files imported by this file. */
  dependency: string[];
  /** Indexes of the public imported files in the dependency list above. */
  publicDependency: number[];
  /**
   * Indexes of the weak imported files in the dependency list.
   * For Google-internal migration only. Do not use.
   */
  weakDependency: number[];
  /** All top-level definitions in this file. */
  messageType: DescriptorProto[];
  enumType: EnumDescriptorProto[];
  service: ServiceDescriptorProto[];
  extension: FieldDescriptorProto[];
  options?: FileOptions;
  /**
   * This field contains optional information about the original source code.
   * You may safely remove this entire field without harming runtime
   * functionality of the descriptors -- the information is needed only by
   * development tools.
   */
  sourceCodeInfo?: SourceCodeInfo;
  /**
   * The syntax of the proto file.
   * The supported values are "proto2" and "proto3".
   */
  syntax: string;
}

/** Describes a message type. */
export interface DescriptorProto {
  name: string;
  field: FieldDescriptorProto[];
  extension: FieldDescriptorProto[];
  nestedType: DescriptorProto[];
  enumType: EnumDescriptorProto[];
  extensionRange: DescriptorProto_ExtensionRange[];
  oneofDecl: OneofDescriptorProto[];
  options?: MessageOptions;
  reservedRange: DescriptorProto_ReservedRange[];
  /**
   * Reserved field names, which may not be used by fields in the same message.
   * A given name may only be reserved once.
   */
  reservedName: string[];
}

export interface DescriptorProto_ExtensionRange {
  /** Inclusive. */
  start: number;
  /** Exclusive. */
  end: number;
  options?: ExtensionRangeOptions;
}

/**
 * Range of reserved tag numbers. Reserved tag numbers may not be used by
 * fields or extension ranges in the same message. Reserved ranges may
 * not overlap.
 */
export interface DescriptorProto_ReservedRange {
  /** Inclusive. */
  start: number;
  /** Exclusive. */
  end: number;
}

export interface ExtensionRangeOptions {
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

/** Describes a field within a message. */
export interface FieldDescriptorProto {
  name: string;
  number: number;
  label: FieldDescriptorProto_Label;
  /**
   * If type_name is set, this need not be set.  If both this and type_name
   * are set, this must be one of TYPE_ENUM, TYPE_MESSAGE or TYPE_GROUP.
   */
  type: FieldDescriptorProto_Type;
  /**
   * For message and enum types, this is the name of the type.  If the name
   * starts with a '.', it is fully-qualified.  Otherwise, C++-like scoping
   * rules are used to find the type (i.e. first the nested types within this
   * message are searched, then within the parent, on up to the root
   * namespace).
   */
  typeName: string;
  /**
   * For extensions, this is the name of the type being extended.  It is
   * resolved in the same manner as type_name.
   */
  extendee: string;
  /**
   * For numeric types, contains the original text representation of the value.
   * For booleans, "true" or "false".
   * For strings, contains the default text contents (not escaped in any way).
   * For bytes, contains the C escaped value.  All bytes >= 128 are escaped.
   * TODO(kenton):  Base-64 encode?
   */
  defaultValue: string;
  /**
   * If set, gives the index of a oneof in the containing type's oneof_decl
   * list.  This field is a member of that oneof.
   */
  oneofIndex: number;
  /**
   * JSON name of this field. The value is set by protocol compiler. If the
   * user has set a "json_name" option on this field, that option's value
   * will be used. Otherwise, it's deduced from the field's name by converting
   * it to camelCase.
   */
  jsonName: string;
  options?: FieldOptions;
  /**
   * If true, this is a proto3 "optional". When a proto3 field is optional, it
   * tracks presence regardless of field type.
   *
   * When proto3_optional is true, this field must be belong to a oneof to
   * signal to old proto3 clients that presence is tracked for this field. This
   * oneof is known as a "synthetic" oneof, and this field must be its sole
   * member (each proto3 optional field gets its own synthetic oneof). Synthetic
   * oneofs exist in the descriptor only, and do not generate any API. Synthetic
   * oneofs must be ordered after all "real" oneofs.
   *
   * For message fields, proto3_optional doesn't create any semantic change,
   * since non-repeated message fields always track presence. However it still
   * indicates the semantic detail of whether the user wrote "optional" or not.
   * This can be useful for round-tripping the .proto file. For consistency we
   * give message fields a synthetic oneof also, even though it is not required
   * to track presence. This is especially important because the parser can't
   * tell if a field is a message or an enum, so it must always create a
   * synthetic oneof.
   *
   * Proto2 optional fields do not set this flag, because they already indicate
   * optional with `LABEL_OPTIONAL`.
   */
  proto3Optional: boolean;
}

export enum FieldDescriptorProto_Type {
  /**
   * TYPE_DOUBLE - 0 is reserved for errors.
   * Order is weird for historical reasons.
   */
  TYPE_DOUBLE = 1,
  TYPE_FLOAT = 2,
  /**
   * TYPE_INT64 - Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
   * negative values are likely.
   */
  TYPE_INT64 = 3,
  TYPE_UINT64 = 4,
  /**
   * TYPE_INT32 - Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
   * negative values are likely.
   */
  TYPE_INT32 = 5,
  TYPE_FIXED64 = 6,
  TYPE_FIXED32 = 7,
  TYPE_BOOL = 8,
  TYPE_STRING = 9,
  /**
   * TYPE_GROUP - Tag-delimited aggregate.
   * Group type is deprecated and not supported in proto3. However, Proto3
   * implementations should still be able to parse the group wire format and
   * treat group fields as unknown fields.
   */
  TYPE_GROUP = 10,
  /** TYPE_MESSAGE - Length-delimited aggregate. */
  TYPE_MESSAGE = 11,
  /** TYPE_BYTES - New in version 2. */
  TYPE_BYTES = 12,
  TYPE_UINT32 = 13,
  TYPE_ENUM = 14,
  TYPE_SFIXED32 = 15,
  TYPE_SFIXED64 = 16,
  /** TYPE_SINT32 - Uses ZigZag encoding. */
  TYPE_SINT32 = 17,
  /** TYPE_SINT64 - Uses ZigZag encoding. */
  TYPE_SINT64 = 18,
  UNRECOGNIZED = -1,
}

export function fieldDescriptorProto_TypeFromJSON(
  object: any
): FieldDescriptorProto_Type {
  switch (object) {
    case 1:
    case "TYPE_DOUBLE":
      return FieldDescriptorProto_Type.TYPE_DOUBLE;
    case 2:
    case "TYPE_FLOAT":
      return FieldDescriptorProto_Type.TYPE_FLOAT;
    case 3:
    case "TYPE_INT64":
      return FieldDescriptorProto_Type.TYPE_INT64;
    case 4:
    case "TYPE_UINT64":
      return FieldDescriptorProto_Type.TYPE_UINT64;
    case 5:
    case "TYPE_INT32":
      return FieldDescriptorProto_Type.TYPE_INT32;
    case 6:
    case "TYPE_FIXED64":
      return FieldDescriptorProto_Type.TYPE_FIXED64;
    case 7:
    case "TYPE_FIXED32":
      return FieldDescriptorProto_Type.TYPE_FIXED32;
    case 8:
    case "TYPE_BOOL":
      return FieldDescriptorProto_Type.TYPE_BOOL;
    case 9:
    case "TYPE_STRING":
      return FieldDescriptorProto_Type.TYPE_STRING;
    case 10:
    case "TYPE_GROUP":
      return FieldDescriptorProto_Type.TYPE_GROUP;
    case 11:
    case "TYPE_MESSAGE":
      return FieldDescriptorProto_Type.TYPE_MESSAGE;
    case 12:
    case "TYPE_BYTES":
      return FieldDescriptorProto_Type.TYPE_BYTES;
    case 13:
    case "TYPE_UINT32":
      return FieldDescriptorProto_Type.TYPE_UINT32;
    case 14:
    case "TYPE_ENUM":
      return FieldDescriptorProto_Type.TYPE_ENUM;
    case 15:
    case "TYPE_SFIXED32":
      return FieldDescriptorProto_Type.TYPE_SFIXED32;
    case 16:
    case "TYPE_SFIXED64":
      return FieldDescriptorProto_Type.TYPE_SFIXED64;
    case 17:
    case "TYPE_SINT32":
      return FieldDescriptorProto_Type.TYPE_SINT32;
    case 18:
    case "TYPE_SINT64":
      return FieldDescriptorProto_Type.TYPE_SINT64;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FieldDescriptorProto_Type.UNRECOGNIZED;
  }
}

export function fieldDescriptorProto_TypeToJSON(
  object: FieldDescriptorProto_Type
): string {
  switch (object) {
    case FieldDescriptorProto_Type.TYPE_DOUBLE:
      return "TYPE_DOUBLE";
    case FieldDescriptorProto_Type.TYPE_FLOAT:
      return "TYPE_FLOAT";
    case FieldDescriptorProto_Type.TYPE_INT64:
      return "TYPE_INT64";
    case FieldDescriptorProto_Type.TYPE_UINT64:
      return "TYPE_UINT64";
    case FieldDescriptorProto_Type.TYPE_INT32:
      return "TYPE_INT32";
    case FieldDescriptorProto_Type.TYPE_FIXED64:
      return "TYPE_FIXED64";
    case FieldDescriptorProto_Type.TYPE_FIXED32:
      return "TYPE_FIXED32";
    case FieldDescriptorProto_Type.TYPE_BOOL:
      return "TYPE_BOOL";
    case FieldDescriptorProto_Type.TYPE_STRING:
      return "TYPE_STRING";
    case FieldDescriptorProto_Type.TYPE_GROUP:
      return "TYPE_GROUP";
    case FieldDescriptorProto_Type.TYPE_MESSAGE:
      return "TYPE_MESSAGE";
    case FieldDescriptorProto_Type.TYPE_BYTES:
      return "TYPE_BYTES";
    case FieldDescriptorProto_Type.TYPE_UINT32:
      return "TYPE_UINT32";
    case FieldDescriptorProto_Type.TYPE_ENUM:
      return "TYPE_ENUM";
    case FieldDescriptorProto_Type.TYPE_SFIXED32:
      return "TYPE_SFIXED32";
    case FieldDescriptorProto_Type.TYPE_SFIXED64:
      return "TYPE_SFIXED64";
    case FieldDescriptorProto_Type.TYPE_SINT32:
      return "TYPE_SINT32";
    case FieldDescriptorProto_Type.TYPE_SINT64:
      return "TYPE_SINT64";
    default:
      return "UNKNOWN";
  }
}

export enum FieldDescriptorProto_Label {
  /** LABEL_OPTIONAL - 0 is reserved for errors */
  LABEL_OPTIONAL = 1,
  LABEL_REQUIRED = 2,
  LABEL_REPEATED = 3,
  UNRECOGNIZED = -1,
}

export function fieldDescriptorProto_LabelFromJSON(
  object: any
): FieldDescriptorProto_Label {
  switch (object) {
    case 1:
    case "LABEL_OPTIONAL":
      return FieldDescriptorProto_Label.LABEL_OPTIONAL;
    case 2:
    case "LABEL_REQUIRED":
      return FieldDescriptorProto_Label.LABEL_REQUIRED;
    case 3:
    case "LABEL_REPEATED":
      return FieldDescriptorProto_Label.LABEL_REPEATED;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FieldDescriptorProto_Label.UNRECOGNIZED;
  }
}

export function fieldDescriptorProto_LabelToJSON(
  object: FieldDescriptorProto_Label
): string {
  switch (object) {
    case FieldDescriptorProto_Label.LABEL_OPTIONAL:
      return "LABEL_OPTIONAL";
    case FieldDescriptorProto_Label.LABEL_REQUIRED:
      return "LABEL_REQUIRED";
    case FieldDescriptorProto_Label.LABEL_REPEATED:
      return "LABEL_REPEATED";
    default:
      return "UNKNOWN";
  }
}

/** Describes a oneof. */
export interface OneofDescriptorProto {
  name: string;
  options?: OneofOptions;
}

/** Describes an enum type. */
export interface EnumDescriptorProto {
  name: string;
  value: EnumValueDescriptorProto[];
  options?: EnumOptions;
  /**
   * Range of reserved numeric values. Reserved numeric values may not be used
   * by enum values in the same enum declaration. Reserved ranges may not
   * overlap.
   */
  reservedRange: EnumDescriptorProto_EnumReservedRange[];
  /**
   * Reserved enum value names, which may not be reused. A given name may only
   * be reserved once.
   */
  reservedName: string[];
}

/**
 * Range of reserved numeric values. Reserved values may not be used by
 * entries in the same enum. Reserved ranges may not overlap.
 *
 * Note that this is distinct from DescriptorProto.ReservedRange in that it
 * is inclusive such that it can appropriately represent the entire int32
 * domain.
 */
export interface EnumDescriptorProto_EnumReservedRange {
  /** Inclusive. */
  start: number;
  /** Inclusive. */
  end: number;
}

/** Describes a value within an enum. */
export interface EnumValueDescriptorProto {
  name: string;
  number: number;
  options?: EnumValueOptions;
}

/** Describes a service. */
export interface ServiceDescriptorProto {
  name: string;
  method: MethodDescriptorProto[];
  options?: ServiceOptions;
}

/** Describes a method of a service. */
export interface MethodDescriptorProto {
  name: string;
  /**
   * Input and output type names.  These are resolved in the same way as
   * FieldDescriptorProto.type_name, but must refer to a message type.
   */
  inputType: string;
  outputType: string;
  options?: MethodOptions;
  /** Identifies if client streams multiple client messages */
  clientStreaming: boolean;
  /** Identifies if server streams multiple server messages */
  serverStreaming: boolean;
}

export interface FileOptions {
  /**
   * Sets the Java package where classes generated from this .proto will be
   * placed.  By default, the proto package is used, but this is often
   * inappropriate because proto packages do not normally start with backwards
   * domain names.
   */
  javaPackage: string;
  /**
   * Controls the name of the wrapper Java class generated for the .proto file.
   * That class will always contain the .proto file's getDescriptor() method as
   * well as any top-level extensions defined in the .proto file.
   * If java_multiple_files is disabled, then all the other classes from the
   * .proto file will be nested inside the single wrapper outer class.
   */
  javaOuterClassname: string;
  /**
   * If enabled, then the Java code generator will generate a separate .java
   * file for each top-level message, enum, and service defined in the .proto
   * file.  Thus, these types will *not* be nested inside the wrapper class
   * named by java_outer_classname.  However, the wrapper class will still be
   * generated to contain the file's getDescriptor() method as well as any
   * top-level extensions defined in the file.
   */
  javaMultipleFiles: boolean;
  /**
   * This option does nothing.
   *
   * @deprecated
   */
  javaGenerateEqualsAndHash: boolean;
  /**
   * If set true, then the Java2 code generator will generate code that
   * throws an exception whenever an attempt is made to assign a non-UTF-8
   * byte sequence to a string field.
   * Message reflection will do the same.
   * However, an extension field still accepts non-UTF-8 byte sequences.
   * This option has no effect on when used with the lite runtime.
   */
  javaStringCheckUtf8: boolean;
  optimizeFor: FileOptions_OptimizeMode;
  /**
   * Sets the Go package where structs generated from this .proto will be
   * placed. If omitted, the Go package will be derived from the following:
   *   - The basename of the package import path, if provided.
   *   - Otherwise, the package statement in the .proto file, if present.
   *   - Otherwise, the basename of the .proto file, without extension.
   */
  goPackage: string;
  /**
   * Should generic services be generated in each language?  "Generic" services
   * are not specific to any particular RPC system.  They are generated by the
   * main code generators in each language (without additional plugins).
   * Generic services were the only kind of service generation supported by
   * early versions of google.protobuf.
   *
   * Generic services are now considered deprecated in favor of using plugins
   * that generate code specific to your particular RPC system.  Therefore,
   * these default to false.  Old code which depends on generic services should
   * explicitly set them to true.
   */
  ccGenericServices: boolean;
  javaGenericServices: boolean;
  pyGenericServices: boolean;
  phpGenericServices: boolean;
  /**
   * Is this file deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for everything in the file, or it will be completely ignored; in the very
   * least, this is a formalization for deprecating files.
   */
  deprecated: boolean;
  /**
   * Enables the use of arenas for the proto messages in this file. This applies
   * only to generated classes for C++.
   */
  ccEnableArenas: boolean;
  /**
   * Sets the objective c class prefix which is prepended to all objective c
   * generated classes from this .proto. There is no default.
   */
  objcClassPrefix: string;
  /** Namespace for generated classes; defaults to the package. */
  csharpNamespace: string;
  /**
   * By default Swift generators will take the proto package and CamelCase it
   * replacing '.' with underscore and use that to prefix the types/symbols
   * defined. When this options is provided, they will use this value instead
   * to prefix the types/symbols defined.
   */
  swiftPrefix: string;
  /**
   * Sets the php class prefix which is prepended to all php generated classes
   * from this .proto. Default is empty.
   */
  phpClassPrefix: string;
  /**
   * Use this option to change the namespace of php generated classes. Default
   * is empty. When this option is empty, the package name will be used for
   * determining the namespace.
   */
  phpNamespace: string;
  /**
   * Use this option to change the namespace of php generated metadata classes.
   * Default is empty. When this option is empty, the proto file name will be
   * used for determining the namespace.
   */
  phpMetadataNamespace: string;
  /**
   * Use this option to change the package of ruby generated classes. Default
   * is empty. When this option is not set, the package name will be used for
   * determining the ruby package.
   */
  rubyPackage: string;
  /**
   * The parser stores options it doesn't recognize here.
   * See the documentation for the "Options" section above.
   */
  uninterpretedOption: UninterpretedOption[];
}

/** Generated classes can be optimized for speed or code size. */
export enum FileOptions_OptimizeMode {
  /** SPEED - Generate complete code for parsing, serialization, */
  SPEED = 1,
  /** CODE_SIZE - etc. */
  CODE_SIZE = 2,
  /** LITE_RUNTIME - Generate code using MessageLite and the lite runtime. */
  LITE_RUNTIME = 3,
  UNRECOGNIZED = -1,
}

export function fileOptions_OptimizeModeFromJSON(
  object: any
): FileOptions_OptimizeMode {
  switch (object) {
    case 1:
    case "SPEED":
      return FileOptions_OptimizeMode.SPEED;
    case 2:
    case "CODE_SIZE":
      return FileOptions_OptimizeMode.CODE_SIZE;
    case 3:
    case "LITE_RUNTIME":
      return FileOptions_OptimizeMode.LITE_RUNTIME;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FileOptions_OptimizeMode.UNRECOGNIZED;
  }
}

export function fileOptions_OptimizeModeToJSON(
  object: FileOptions_OptimizeMode
): string {
  switch (object) {
    case FileOptions_OptimizeMode.SPEED:
      return "SPEED";
    case FileOptions_OptimizeMode.CODE_SIZE:
      return "CODE_SIZE";
    case FileOptions_OptimizeMode.LITE_RUNTIME:
      return "LITE_RUNTIME";
    default:
      return "UNKNOWN";
  }
}

export interface MessageOptions {
  /**
   * Set true to use the old proto1 MessageSet wire format for extensions.
   * This is provided for backwards-compatibility with the MessageSet wire
   * format.  You should not use this for any other reason:  It's less
   * efficient, has fewer features, and is more complicated.
   *
   * The message must be defined exactly as follows:
   *   message Foo {
   *     option message_set_wire_format = true;
   *     extensions 4 to max;
   *   }
   * Note that the message cannot have any defined fields; MessageSets only
   * have extensions.
   *
   * All extensions of your type must be singular messages; e.g. they cannot
   * be int32s, enums, or repeated messages.
   *
   * Because this is an option, the above two restrictions are not enforced by
   * the protocol compiler.
   */
  messageSetWireFormat: boolean;
  /**
   * Disables the generation of the standard "descriptor()" accessor, which can
   * conflict with a field of the same name.  This is meant to make migration
   * from proto1 easier; new code should avoid fields named "descriptor".
   */
  noStandardDescriptorAccessor: boolean;
  /**
   * Is this message deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for the message, or it will be completely ignored; in the very least,
   * this is a formalization for deprecating messages.
   */
  deprecated: boolean;
  /**
   * Whether the message is an automatically generated map entry type for the
   * maps field.
   *
   * For maps fields:
   *     map<KeyType, ValueType> map_field = 1;
   * The parsed descriptor looks like:
   *     message MapFieldEntry {
   *         option map_entry = true;
   *         optional KeyType key = 1;
   *         optional ValueType value = 2;
   *     }
   *     repeated MapFieldEntry map_field = 1;
   *
   * Implementations may choose not to generate the map_entry=true message, but
   * use a native map in the target language to hold the keys and values.
   * The reflection APIs in such implementations still need to work as
   * if the field is a repeated message field.
   *
   * NOTE: Do not set the option in .proto files. Always use the maps syntax
   * instead. The option should only be implicitly set by the proto compiler
   * parser.
   */
  mapEntry: boolean;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export interface FieldOptions {
  /**
   * The ctype option instructs the C++ code generator to use a different
   * representation of the field than it normally would.  See the specific
   * options below.  This option is not yet implemented in the open source
   * release -- sorry, we'll try to include it in a future version!
   */
  ctype: FieldOptions_CType;
  /**
   * The packed option can be enabled for repeated primitive fields to enable
   * a more efficient representation on the wire. Rather than repeatedly
   * writing the tag and type for each element, the entire array is encoded as
   * a single length-delimited blob. In proto3, only explicit setting it to
   * false will avoid using packed encoding.
   */
  packed: boolean;
  /**
   * The jstype option determines the JavaScript type used for values of the
   * field.  The option is permitted only for 64 bit integral and fixed types
   * (int64, uint64, sint64, fixed64, sfixed64).  A field with jstype JS_STRING
   * is represented as JavaScript string, which avoids loss of precision that
   * can happen when a large value is converted to a floating point JavaScript.
   * Specifying JS_NUMBER for the jstype causes the generated JavaScript code to
   * use the JavaScript "number" type.  The behavior of the default option
   * JS_NORMAL is implementation dependent.
   *
   * This option is an enum to permit additional types to be added, e.g.
   * goog.math.Integer.
   */
  jstype: FieldOptions_JSType;
  /**
   * Should this field be parsed lazily?  Lazy applies only to message-type
   * fields.  It means that when the outer message is initially parsed, the
   * inner message's contents will not be parsed but instead stored in encoded
   * form.  The inner message will actually be parsed when it is first accessed.
   *
   * This is only a hint.  Implementations are free to choose whether to use
   * eager or lazy parsing regardless of the value of this option.  However,
   * setting this option true suggests that the protocol author believes that
   * using lazy parsing on this field is worth the additional bookkeeping
   * overhead typically needed to implement it.
   *
   * This option does not affect the public interface of any generated code;
   * all method signatures remain the same.  Furthermore, thread-safety of the
   * interface is not affected by this option; const methods remain safe to
   * call from multiple threads concurrently, while non-const methods continue
   * to require exclusive access.
   *
   *
   * Note that implementations may choose not to check required fields within
   * a lazy sub-message.  That is, calling IsInitialized() on the outer message
   * may return true even if the inner message has missing required fields.
   * This is necessary because otherwise the inner message would have to be
   * parsed in order to perform the check, defeating the purpose of lazy
   * parsing.  An implementation which chooses not to check required fields
   * must be consistent about it.  That is, for any particular sub-message, the
   * implementation must either *always* check its required fields, or *never*
   * check its required fields, regardless of whether or not the message has
   * been parsed.
   */
  lazy: boolean;
  /**
   * Is this field deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for accessors, or it will be completely ignored; in the very least, this
   * is a formalization for deprecating fields.
   */
  deprecated: boolean;
  /** For Google-internal migration only. Do not use. */
  weak: boolean;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export enum FieldOptions_CType {
  /** STRING - Default mode. */
  STRING = 0,
  CORD = 1,
  STRING_PIECE = 2,
  UNRECOGNIZED = -1,
}

export function fieldOptions_CTypeFromJSON(object: any): FieldOptions_CType {
  switch (object) {
    case 0:
    case "STRING":
      return FieldOptions_CType.STRING;
    case 1:
    case "CORD":
      return FieldOptions_CType.CORD;
    case 2:
    case "STRING_PIECE":
      return FieldOptions_CType.STRING_PIECE;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FieldOptions_CType.UNRECOGNIZED;
  }
}

export function fieldOptions_CTypeToJSON(object: FieldOptions_CType): string {
  switch (object) {
    case FieldOptions_CType.STRING:
      return "STRING";
    case FieldOptions_CType.CORD:
      return "CORD";
    case FieldOptions_CType.STRING_PIECE:
      return "STRING_PIECE";
    default:
      return "UNKNOWN";
  }
}

export enum FieldOptions_JSType {
  /** JS_NORMAL - Use the default type. */
  JS_NORMAL = 0,
  /** JS_STRING - Use JavaScript strings. */
  JS_STRING = 1,
  /** JS_NUMBER - Use JavaScript numbers. */
  JS_NUMBER = 2,
  UNRECOGNIZED = -1,
}

export function fieldOptions_JSTypeFromJSON(object: any): FieldOptions_JSType {
  switch (object) {
    case 0:
    case "JS_NORMAL":
      return FieldOptions_JSType.JS_NORMAL;
    case 1:
    case "JS_STRING":
      return FieldOptions_JSType.JS_STRING;
    case 2:
    case "JS_NUMBER":
      return FieldOptions_JSType.JS_NUMBER;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FieldOptions_JSType.UNRECOGNIZED;
  }
}

export function fieldOptions_JSTypeToJSON(object: FieldOptions_JSType): string {
  switch (object) {
    case FieldOptions_JSType.JS_NORMAL:
      return "JS_NORMAL";
    case FieldOptions_JSType.JS_STRING:
      return "JS_STRING";
    case FieldOptions_JSType.JS_NUMBER:
      return "JS_NUMBER";
    default:
      return "UNKNOWN";
  }
}

export interface OneofOptions {
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export interface EnumOptions {
  /**
   * Set this option to true to allow mapping different tag names to the same
   * value.
   */
  allowAlias: boolean;
  /**
   * Is this enum deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for the enum, or it will be completely ignored; in the very least, this
   * is a formalization for deprecating enums.
   */
  deprecated: boolean;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export interface EnumValueOptions {
  /**
   * Is this enum value deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for the enum value, or it will be completely ignored; in the very least,
   * this is a formalization for deprecating enum values.
   */
  deprecated: boolean;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export interface ServiceOptions {
  /**
   * Is this service deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for the service, or it will be completely ignored; in the very least,
   * this is a formalization for deprecating services.
   */
  deprecated: boolean;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

export interface MethodOptions {
  /**
   * Is this method deprecated?
   * Depending on the target platform, this can emit Deprecated annotations
   * for the method, or it will be completely ignored; in the very least,
   * this is a formalization for deprecating methods.
   */
  deprecated: boolean;
  idempotencyLevel: MethodOptions_IdempotencyLevel;
  /** The parser stores options it doesn't recognize here. See above. */
  uninterpretedOption: UninterpretedOption[];
}

/**
 * Is this method side-effect-free (or safe in HTTP parlance), or idempotent,
 * or neither? HTTP based RPC implementation may choose GET verb for safe
 * methods, and PUT verb for idempotent methods instead of the default POST.
 */
export enum MethodOptions_IdempotencyLevel {
  IDEMPOTENCY_UNKNOWN = 0,
  /** NO_SIDE_EFFECTS - implies idempotent */
  NO_SIDE_EFFECTS = 1,
  /** IDEMPOTENT - idempotent, but may have side effects */
  IDEMPOTENT = 2,
  UNRECOGNIZED = -1,
}

export function methodOptions_IdempotencyLevelFromJSON(
  object: any
): MethodOptions_IdempotencyLevel {
  switch (object) {
    case 0:
    case "IDEMPOTENCY_UNKNOWN":
      return MethodOptions_IdempotencyLevel.IDEMPOTENCY_UNKNOWN;
    case 1:
    case "NO_SIDE_EFFECTS":
      return MethodOptions_IdempotencyLevel.NO_SIDE_EFFECTS;
    case 2:
    case "IDEMPOTENT":
      return MethodOptions_IdempotencyLevel.IDEMPOTENT;
    case -1:
    case "UNRECOGNIZED":
    default:
      return MethodOptions_IdempotencyLevel.UNRECOGNIZED;
  }
}

export function methodOptions_IdempotencyLevelToJSON(
  object: MethodOptions_IdempotencyLevel
): string {
  switch (object) {
    case MethodOptions_IdempotencyLevel.IDEMPOTENCY_UNKNOWN:
      return "IDEMPOTENCY_UNKNOWN";
    case MethodOptions_IdempotencyLevel.NO_SIDE_EFFECTS:
      return "NO_SIDE_EFFECTS";
    case MethodOptions_IdempotencyLevel.IDEMPOTENT:
      return "IDEMPOTENT";
    default:
      return "UNKNOWN";
  }
}

/**
 * A message representing a option the parser does not recognize. This only
 * appears in options protos created by the compiler::Parser class.
 * DescriptorPool resolves these when building Descriptor objects. Therefore,
 * options protos in descriptor objects (e.g. returned by Descriptor::options(),
 * or produced by Descriptor::CopyTo()) will never have UninterpretedOptions
 * in them.
 */
export interface UninterpretedOption {
  name: UninterpretedOption_NamePart[];
  /**
   * The value of the uninterpreted option, in whatever type the tokenizer
   * identified it as during parsing. Exactly one of these should be set.
   */
  identifierValue: string;
  positiveIntValue: number;
  negativeIntValue: number;
  doubleValue: number;
  stringValue: Uint8Array;
  aggregateValue: string;
}

/**
 * The name of the uninterpreted option.  Each string represents a segment in
 * a dot-separated name.  is_extension is true iff a segment represents an
 * extension (denoted with parentheses in options specs in .proto files).
 * E.g.,{ ["foo", false], ["bar.baz", true], ["qux", false] } represents
 * "foo.(bar.baz).qux".
 */
export interface UninterpretedOption_NamePart {
  namePart: string;
  isExtension: boolean;
}

/**
 * Encapsulates information about the original source file from which a
 * FileDescriptorProto was generated.
 */
export interface SourceCodeInfo {
  /**
   * A Location identifies a piece of source code in a .proto file which
   * corresponds to a particular definition.  This information is intended
   * to be useful to IDEs, code indexers, documentation generators, and similar
   * tools.
   *
   * For example, say we have a file like:
   *   message Foo {
   *     optional string foo = 1;
   *   }
   * Let's look at just the field definition:
   *   optional string foo = 1;
   *   ^       ^^     ^^  ^  ^^^
   *   a       bc     de  f  ghi
   * We have the following locations:
   *   span   path               represents
   *   [a,i)  [ 4, 0, 2, 0 ]     The whole field definition.
   *   [a,b)  [ 4, 0, 2, 0, 4 ]  The label (optional).
   *   [c,d)  [ 4, 0, 2, 0, 5 ]  The type (string).
   *   [e,f)  [ 4, 0, 2, 0, 1 ]  The name (foo).
   *   [g,h)  [ 4, 0, 2, 0, 3 ]  The number (1).
   *
   * Notes:
   * - A location may refer to a repeated field itself (i.e. not to any
   *   particular index within it).  This is used whenever a set of elements are
   *   logically enclosed in a single code segment.  For example, an entire
   *   extend block (possibly containing multiple extension definitions) will
   *   have an outer location whose path refers to the "extensions" repeated
   *   field without an index.
   * - Multiple locations may have the same path.  This happens when a single
   *   logical declaration is spread out across multiple places.  The most
   *   obvious example is the "extend" block again -- there may be multiple
   *   extend blocks in the same scope, each of which will have the same path.
   * - A location's span is not always a subset of its parent's span.  For
   *   example, the "extendee" of an extension declaration appears at the
   *   beginning of the "extend" block and is shared by all extensions within
   *   the block.
   * - Just because a location's span is a subset of some other location's span
   *   does not mean that it is a descendant.  For example, a "group" defines
   *   both a type and a field in a single declaration.  Thus, the locations
   *   corresponding to the type and field and their components will overlap.
   * - Code which tries to interpret locations should probably be designed to
   *   ignore those that it doesn't understand, as more types of locations could
   *   be recorded in the future.
   */
  location: SourceCodeInfo_Location[];
}

export interface SourceCodeInfo_Location {
  /**
   * Identifies which part of the FileDescriptorProto was defined at this
   * location.
   *
   * Each element is a field number or an index.  They form a path from
   * the root FileDescriptorProto to the place where the definition.  For
   * example, this path:
   *   [ 4, 3, 2, 7, 1 ]
   * refers to:
   *   file.message_type(3)  // 4, 3
   *       .field(7)         // 2, 7
   *       .name()           // 1
   * This is because FileDescriptorProto.message_type has field number 4:
   *   repeated DescriptorProto message_type = 4;
   * and DescriptorProto.field has field number 2:
   *   repeated FieldDescriptorProto field = 2;
   * and FieldDescriptorProto.name has field number 1:
   *   optional string name = 1;
   *
   * Thus, the above path gives the location of a field name.  If we removed
   * the last element:
   *   [ 4, 3, 2, 7 ]
   * this path refers to the whole field declaration (from the beginning
   * of the label to the terminating semicolon).
   */
  path: number[];
  /**
   * Always has exactly three or four elements: start line, start column,
   * end line (optional, otherwise assumed same as start line), end column.
   * These are packed into a single field for efficiency.  Note that line
   * and column numbers are zero-based -- typically you will want to add
   * 1 to each before displaying to a user.
   */
  span: number[];
  /**
   * If this SourceCodeInfo represents a complete declaration, these are any
   * comments appearing before and after the declaration which appear to be
   * attached to the declaration.
   *
   * A series of line comments appearing on consecutive lines, with no other
   * tokens appearing on those lines, will be treated as a single comment.
   *
   * leading_detached_comments will keep paragraphs of comments that appear
   * before (but not connected to) the current element. Each paragraph,
   * separated by empty lines, will be one comment element in the repeated
   * field.
   *
   * Only the comment content is provided; comment markers (e.g. //) are
   * stripped out.  For block comments, leading whitespace and an asterisk
   * will be stripped from the beginning of each line other than the first.
   * Newlines are included in the output.
   *
   * Examples:
   *
   *   optional int32 foo = 1;  // Comment attached to foo.
   *   // Comment attached to bar.
   *   optional int32 bar = 2;
   *
   *   optional string baz = 3;
   *   // Comment attached to baz.
   *   // Another line attached to baz.
   *
   *   // Comment attached to qux.
   *   //
   *   // Another line attached to qux.
   *   optional double qux = 4;
   *
   *   // Detached comment for corge. This is not leading or trailing comments
   *   // to qux or corge because there are blank lines separating it from
   *   // both.
   *
   *   // Detached comment for corge paragraph 2.
   *
   *   optional string corge = 5;
   *   /* Block comment attached
   *    * to corge.  Leading asterisks
   *    * will be removed. * /
   *   /* Block comment attached to
   *    * grault. * /
   *   optional int32 grault = 6;
   *
   *   // ignored detached comments.
   */
  leadingComments: string;
  trailingComments: string;
  leadingDetachedComments: string[];
}

/**
 * Describes the relationship between generated code and its original source
 * file. A GeneratedCodeInfo message is associated with only one generated
 * source file, but may contain references to different source .proto files.
 */
export interface GeneratedCodeInfo {
  /**
   * An Annotation connects some span of text in generated code to an element
   * of its generating .proto file.
   */
  annotation: GeneratedCodeInfo_Annotation[];
}

export interface GeneratedCodeInfo_Annotation {
  /**
   * Identifies the element in the original source .proto file. This field
   * is formatted the same as SourceCodeInfo.Location.path.
   */
  path: number[];
  /** Identifies the filesystem path to the original source .proto. */
  sourceFile: string;
  /**
   * Identifies the starting offset in bytes in the generated code
   * that relates to the identified object.
   */
  begin: number;
  /**
   * Identifies the ending offset in bytes in the generated code that
   * relates to the identified offset. The end offset should be one past
   * the last relevant byte (so the length of the text = end - begin).
   */
  end: number;
}

const baseFileDescriptorSet: object = {};

export const FileDescriptorSet = {
  encode(
    message: FileDescriptorSet,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.file) {
      FileDescriptorProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileDescriptorSet {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFileDescriptorSet } as FileDescriptorSet;
    message.file = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.file.push(
            FileDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FileDescriptorSet {
    const message = { ...baseFileDescriptorSet } as FileDescriptorSet;
    message.file = [];
    if (object.file !== undefined && object.file !== null) {
      for (const e of object.file) {
        message.file.push(FileDescriptorProto.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: FileDescriptorSet): unknown {
    const obj: any = {};
    if (message.file) {
      obj.file = message.file.map((e) =>
        e ? FileDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.file = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<FileDescriptorSet>): FileDescriptorSet {
    const message = { ...baseFileDescriptorSet } as FileDescriptorSet;
    message.file = [];
    if (object.file !== undefined && object.file !== null) {
      for (const e of object.file) {
        message.file.push(FileDescriptorProto.fromPartial(e));
      }
    }
    return message;
  },
};

const baseFileDescriptorProto: object = {
  name: "",
  package: "",
  dependency: "",
  publicDependency: 0,
  weakDependency: 0,
  syntax: "",
};

export const FileDescriptorProto = {
  encode(
    message: FileDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.package !== "") {
      writer.uint32(18).string(message.package);
    }
    for (const v of message.dependency) {
      writer.uint32(26).string(v!);
    }
    writer.uint32(82).fork();
    for (const v of message.publicDependency) {
      writer.int32(v);
    }
    writer.ldelim();
    writer.uint32(90).fork();
    for (const v of message.weakDependency) {
      writer.int32(v);
    }
    writer.ldelim();
    for (const v of message.messageType) {
      DescriptorProto.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    for (const v of message.enumType) {
      EnumDescriptorProto.encode(v!, writer.uint32(42).fork()).ldelim();
    }
    for (const v of message.service) {
      ServiceDescriptorProto.encode(v!, writer.uint32(50).fork()).ldelim();
    }
    for (const v of message.extension) {
      FieldDescriptorProto.encode(v!, writer.uint32(58).fork()).ldelim();
    }
    if (message.options !== undefined) {
      FileOptions.encode(message.options, writer.uint32(66).fork()).ldelim();
    }
    if (message.sourceCodeInfo !== undefined) {
      SourceCodeInfo.encode(
        message.sourceCodeInfo,
        writer.uint32(74).fork()
      ).ldelim();
    }
    if (message.syntax !== "") {
      writer.uint32(98).string(message.syntax);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFileDescriptorProto } as FileDescriptorProto;
    message.dependency = [];
    message.publicDependency = [];
    message.weakDependency = [];
    message.messageType = [];
    message.enumType = [];
    message.service = [];
    message.extension = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.package = reader.string();
          break;
        case 3:
          message.dependency.push(reader.string());
          break;
        case 10:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.publicDependency.push(reader.int32());
            }
          } else {
            message.publicDependency.push(reader.int32());
          }
          break;
        case 11:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.weakDependency.push(reader.int32());
            }
          } else {
            message.weakDependency.push(reader.int32());
          }
          break;
        case 4:
          message.messageType.push(
            DescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 5:
          message.enumType.push(
            EnumDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 6:
          message.service.push(
            ServiceDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 7:
          message.extension.push(
            FieldDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 8:
          message.options = FileOptions.decode(reader, reader.uint32());
          break;
        case 9:
          message.sourceCodeInfo = SourceCodeInfo.decode(
            reader,
            reader.uint32()
          );
          break;
        case 12:
          message.syntax = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FileDescriptorProto {
    const message = { ...baseFileDescriptorProto } as FileDescriptorProto;
    message.dependency = [];
    message.publicDependency = [];
    message.weakDependency = [];
    message.messageType = [];
    message.enumType = [];
    message.service = [];
    message.extension = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.package !== undefined && object.package !== null) {
      message.package = String(object.package);
    } else {
      message.package = "";
    }
    if (object.dependency !== undefined && object.dependency !== null) {
      for (const e of object.dependency) {
        message.dependency.push(String(e));
      }
    }
    if (
      object.publicDependency !== undefined &&
      object.publicDependency !== null
    ) {
      for (const e of object.publicDependency) {
        message.publicDependency.push(Number(e));
      }
    }
    if (object.weakDependency !== undefined && object.weakDependency !== null) {
      for (const e of object.weakDependency) {
        message.weakDependency.push(Number(e));
      }
    }
    if (object.messageType !== undefined && object.messageType !== null) {
      for (const e of object.messageType) {
        message.messageType.push(DescriptorProto.fromJSON(e));
      }
    }
    if (object.enumType !== undefined && object.enumType !== null) {
      for (const e of object.enumType) {
        message.enumType.push(EnumDescriptorProto.fromJSON(e));
      }
    }
    if (object.service !== undefined && object.service !== null) {
      for (const e of object.service) {
        message.service.push(ServiceDescriptorProto.fromJSON(e));
      }
    }
    if (object.extension !== undefined && object.extension !== null) {
      for (const e of object.extension) {
        message.extension.push(FieldDescriptorProto.fromJSON(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = FileOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    if (object.sourceCodeInfo !== undefined && object.sourceCodeInfo !== null) {
      message.sourceCodeInfo = SourceCodeInfo.fromJSON(object.sourceCodeInfo);
    } else {
      message.sourceCodeInfo = undefined;
    }
    if (object.syntax !== undefined && object.syntax !== null) {
      message.syntax = String(object.syntax);
    } else {
      message.syntax = "";
    }
    return message;
  },

  toJSON(message: FileDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.package !== undefined && (obj.package = message.package);
    if (message.dependency) {
      obj.dependency = message.dependency.map((e) => e);
    } else {
      obj.dependency = [];
    }
    if (message.publicDependency) {
      obj.publicDependency = message.publicDependency.map((e) => e);
    } else {
      obj.publicDependency = [];
    }
    if (message.weakDependency) {
      obj.weakDependency = message.weakDependency.map((e) => e);
    } else {
      obj.weakDependency = [];
    }
    if (message.messageType) {
      obj.messageType = message.messageType.map((e) =>
        e ? DescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.messageType = [];
    }
    if (message.enumType) {
      obj.enumType = message.enumType.map((e) =>
        e ? EnumDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.enumType = [];
    }
    if (message.service) {
      obj.service = message.service.map((e) =>
        e ? ServiceDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.service = [];
    }
    if (message.extension) {
      obj.extension = message.extension.map((e) =>
        e ? FieldDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.extension = [];
    }
    message.options !== undefined &&
      (obj.options = message.options
        ? FileOptions.toJSON(message.options)
        : undefined);
    message.sourceCodeInfo !== undefined &&
      (obj.sourceCodeInfo = message.sourceCodeInfo
        ? SourceCodeInfo.toJSON(message.sourceCodeInfo)
        : undefined);
    message.syntax !== undefined && (obj.syntax = message.syntax);
    return obj;
  },

  fromPartial(object: DeepPartial<FileDescriptorProto>): FileDescriptorProto {
    const message = { ...baseFileDescriptorProto } as FileDescriptorProto;
    message.dependency = [];
    message.publicDependency = [];
    message.weakDependency = [];
    message.messageType = [];
    message.enumType = [];
    message.service = [];
    message.extension = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.package !== undefined && object.package !== null) {
      message.package = object.package;
    } else {
      message.package = "";
    }
    if (object.dependency !== undefined && object.dependency !== null) {
      for (const e of object.dependency) {
        message.dependency.push(e);
      }
    }
    if (
      object.publicDependency !== undefined &&
      object.publicDependency !== null
    ) {
      for (const e of object.publicDependency) {
        message.publicDependency.push(e);
      }
    }
    if (object.weakDependency !== undefined && object.weakDependency !== null) {
      for (const e of object.weakDependency) {
        message.weakDependency.push(e);
      }
    }
    if (object.messageType !== undefined && object.messageType !== null) {
      for (const e of object.messageType) {
        message.messageType.push(DescriptorProto.fromPartial(e));
      }
    }
    if (object.enumType !== undefined && object.enumType !== null) {
      for (const e of object.enumType) {
        message.enumType.push(EnumDescriptorProto.fromPartial(e));
      }
    }
    if (object.service !== undefined && object.service !== null) {
      for (const e of object.service) {
        message.service.push(ServiceDescriptorProto.fromPartial(e));
      }
    }
    if (object.extension !== undefined && object.extension !== null) {
      for (const e of object.extension) {
        message.extension.push(FieldDescriptorProto.fromPartial(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = FileOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    if (object.sourceCodeInfo !== undefined && object.sourceCodeInfo !== null) {
      message.sourceCodeInfo = SourceCodeInfo.fromPartial(
        object.sourceCodeInfo
      );
    } else {
      message.sourceCodeInfo = undefined;
    }
    if (object.syntax !== undefined && object.syntax !== null) {
      message.syntax = object.syntax;
    } else {
      message.syntax = "";
    }
    return message;
  },
};

const baseDescriptorProto: object = { name: "", reservedName: "" };

export const DescriptorProto = {
  encode(
    message: DescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    for (const v of message.field) {
      FieldDescriptorProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.extension) {
      FieldDescriptorProto.encode(v!, writer.uint32(50).fork()).ldelim();
    }
    for (const v of message.nestedType) {
      DescriptorProto.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.enumType) {
      EnumDescriptorProto.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    for (const v of message.extensionRange) {
      DescriptorProto_ExtensionRange.encode(
        v!,
        writer.uint32(42).fork()
      ).ldelim();
    }
    for (const v of message.oneofDecl) {
      OneofDescriptorProto.encode(v!, writer.uint32(66).fork()).ldelim();
    }
    if (message.options !== undefined) {
      MessageOptions.encode(message.options, writer.uint32(58).fork()).ldelim();
    }
    for (const v of message.reservedRange) {
      DescriptorProto_ReservedRange.encode(
        v!,
        writer.uint32(74).fork()
      ).ldelim();
    }
    for (const v of message.reservedName) {
      writer.uint32(82).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDescriptorProto } as DescriptorProto;
    message.field = [];
    message.extension = [];
    message.nestedType = [];
    message.enumType = [];
    message.extensionRange = [];
    message.oneofDecl = [];
    message.reservedRange = [];
    message.reservedName = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.field.push(
            FieldDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 6:
          message.extension.push(
            FieldDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.nestedType.push(
            DescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 4:
          message.enumType.push(
            EnumDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 5:
          message.extensionRange.push(
            DescriptorProto_ExtensionRange.decode(reader, reader.uint32())
          );
          break;
        case 8:
          message.oneofDecl.push(
            OneofDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 7:
          message.options = MessageOptions.decode(reader, reader.uint32());
          break;
        case 9:
          message.reservedRange.push(
            DescriptorProto_ReservedRange.decode(reader, reader.uint32())
          );
          break;
        case 10:
          message.reservedName.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DescriptorProto {
    const message = { ...baseDescriptorProto } as DescriptorProto;
    message.field = [];
    message.extension = [];
    message.nestedType = [];
    message.enumType = [];
    message.extensionRange = [];
    message.oneofDecl = [];
    message.reservedRange = [];
    message.reservedName = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.field !== undefined && object.field !== null) {
      for (const e of object.field) {
        message.field.push(FieldDescriptorProto.fromJSON(e));
      }
    }
    if (object.extension !== undefined && object.extension !== null) {
      for (const e of object.extension) {
        message.extension.push(FieldDescriptorProto.fromJSON(e));
      }
    }
    if (object.nestedType !== undefined && object.nestedType !== null) {
      for (const e of object.nestedType) {
        message.nestedType.push(DescriptorProto.fromJSON(e));
      }
    }
    if (object.enumType !== undefined && object.enumType !== null) {
      for (const e of object.enumType) {
        message.enumType.push(EnumDescriptorProto.fromJSON(e));
      }
    }
    if (object.extensionRange !== undefined && object.extensionRange !== null) {
      for (const e of object.extensionRange) {
        message.extensionRange.push(DescriptorProto_ExtensionRange.fromJSON(e));
      }
    }
    if (object.oneofDecl !== undefined && object.oneofDecl !== null) {
      for (const e of object.oneofDecl) {
        message.oneofDecl.push(OneofDescriptorProto.fromJSON(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = MessageOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    if (object.reservedRange !== undefined && object.reservedRange !== null) {
      for (const e of object.reservedRange) {
        message.reservedRange.push(DescriptorProto_ReservedRange.fromJSON(e));
      }
    }
    if (object.reservedName !== undefined && object.reservedName !== null) {
      for (const e of object.reservedName) {
        message.reservedName.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: DescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    if (message.field) {
      obj.field = message.field.map((e) =>
        e ? FieldDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.field = [];
    }
    if (message.extension) {
      obj.extension = message.extension.map((e) =>
        e ? FieldDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.extension = [];
    }
    if (message.nestedType) {
      obj.nestedType = message.nestedType.map((e) =>
        e ? DescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.nestedType = [];
    }
    if (message.enumType) {
      obj.enumType = message.enumType.map((e) =>
        e ? EnumDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.enumType = [];
    }
    if (message.extensionRange) {
      obj.extensionRange = message.extensionRange.map((e) =>
        e ? DescriptorProto_ExtensionRange.toJSON(e) : undefined
      );
    } else {
      obj.extensionRange = [];
    }
    if (message.oneofDecl) {
      obj.oneofDecl = message.oneofDecl.map((e) =>
        e ? OneofDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.oneofDecl = [];
    }
    message.options !== undefined &&
      (obj.options = message.options
        ? MessageOptions.toJSON(message.options)
        : undefined);
    if (message.reservedRange) {
      obj.reservedRange = message.reservedRange.map((e) =>
        e ? DescriptorProto_ReservedRange.toJSON(e) : undefined
      );
    } else {
      obj.reservedRange = [];
    }
    if (message.reservedName) {
      obj.reservedName = message.reservedName.map((e) => e);
    } else {
      obj.reservedName = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<DescriptorProto>): DescriptorProto {
    const message = { ...baseDescriptorProto } as DescriptorProto;
    message.field = [];
    message.extension = [];
    message.nestedType = [];
    message.enumType = [];
    message.extensionRange = [];
    message.oneofDecl = [];
    message.reservedRange = [];
    message.reservedName = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.field !== undefined && object.field !== null) {
      for (const e of object.field) {
        message.field.push(FieldDescriptorProto.fromPartial(e));
      }
    }
    if (object.extension !== undefined && object.extension !== null) {
      for (const e of object.extension) {
        message.extension.push(FieldDescriptorProto.fromPartial(e));
      }
    }
    if (object.nestedType !== undefined && object.nestedType !== null) {
      for (const e of object.nestedType) {
        message.nestedType.push(DescriptorProto.fromPartial(e));
      }
    }
    if (object.enumType !== undefined && object.enumType !== null) {
      for (const e of object.enumType) {
        message.enumType.push(EnumDescriptorProto.fromPartial(e));
      }
    }
    if (object.extensionRange !== undefined && object.extensionRange !== null) {
      for (const e of object.extensionRange) {
        message.extensionRange.push(
          DescriptorProto_ExtensionRange.fromPartial(e)
        );
      }
    }
    if (object.oneofDecl !== undefined && object.oneofDecl !== null) {
      for (const e of object.oneofDecl) {
        message.oneofDecl.push(OneofDescriptorProto.fromPartial(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = MessageOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    if (object.reservedRange !== undefined && object.reservedRange !== null) {
      for (const e of object.reservedRange) {
        message.reservedRange.push(
          DescriptorProto_ReservedRange.fromPartial(e)
        );
      }
    }
    if (object.reservedName !== undefined && object.reservedName !== null) {
      for (const e of object.reservedName) {
        message.reservedName.push(e);
      }
    }
    return message;
  },
};

const baseDescriptorProto_ExtensionRange: object = { start: 0, end: 0 };

export const DescriptorProto_ExtensionRange = {
  encode(
    message: DescriptorProto_ExtensionRange,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.start !== 0) {
      writer.uint32(8).int32(message.start);
    }
    if (message.end !== 0) {
      writer.uint32(16).int32(message.end);
    }
    if (message.options !== undefined) {
      ExtensionRangeOptions.encode(
        message.options,
        writer.uint32(26).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): DescriptorProto_ExtensionRange {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDescriptorProto_ExtensionRange,
    } as DescriptorProto_ExtensionRange;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.start = reader.int32();
          break;
        case 2:
          message.end = reader.int32();
          break;
        case 3:
          message.options = ExtensionRangeOptions.decode(
            reader,
            reader.uint32()
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DescriptorProto_ExtensionRange {
    const message = {
      ...baseDescriptorProto_ExtensionRange,
    } as DescriptorProto_ExtensionRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = Number(object.start);
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = Number(object.end);
    } else {
      message.end = 0;
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = ExtensionRangeOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },

  toJSON(message: DescriptorProto_ExtensionRange): unknown {
    const obj: any = {};
    message.start !== undefined && (obj.start = message.start);
    message.end !== undefined && (obj.end = message.end);
    message.options !== undefined &&
      (obj.options = message.options
        ? ExtensionRangeOptions.toJSON(message.options)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<DescriptorProto_ExtensionRange>
  ): DescriptorProto_ExtensionRange {
    const message = {
      ...baseDescriptorProto_ExtensionRange,
    } as DescriptorProto_ExtensionRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = object.start;
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = object.end;
    } else {
      message.end = 0;
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = ExtensionRangeOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },
};

const baseDescriptorProto_ReservedRange: object = { start: 0, end: 0 };

export const DescriptorProto_ReservedRange = {
  encode(
    message: DescriptorProto_ReservedRange,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.start !== 0) {
      writer.uint32(8).int32(message.start);
    }
    if (message.end !== 0) {
      writer.uint32(16).int32(message.end);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): DescriptorProto_ReservedRange {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDescriptorProto_ReservedRange,
    } as DescriptorProto_ReservedRange;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.start = reader.int32();
          break;
        case 2:
          message.end = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DescriptorProto_ReservedRange {
    const message = {
      ...baseDescriptorProto_ReservedRange,
    } as DescriptorProto_ReservedRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = Number(object.start);
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = Number(object.end);
    } else {
      message.end = 0;
    }
    return message;
  },

  toJSON(message: DescriptorProto_ReservedRange): unknown {
    const obj: any = {};
    message.start !== undefined && (obj.start = message.start);
    message.end !== undefined && (obj.end = message.end);
    return obj;
  },

  fromPartial(
    object: DeepPartial<DescriptorProto_ReservedRange>
  ): DescriptorProto_ReservedRange {
    const message = {
      ...baseDescriptorProto_ReservedRange,
    } as DescriptorProto_ReservedRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = object.start;
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = object.end;
    } else {
      message.end = 0;
    }
    return message;
  },
};

const baseExtensionRangeOptions: object = {};

export const ExtensionRangeOptions = {
  encode(
    message: ExtensionRangeOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): ExtensionRangeOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseExtensionRangeOptions } as ExtensionRangeOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ExtensionRangeOptions {
    const message = { ...baseExtensionRangeOptions } as ExtensionRangeOptions;
    message.uninterpretedOption = [];
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: ExtensionRangeOptions): unknown {
    const obj: any = {};
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<ExtensionRangeOptions>
  ): ExtensionRangeOptions {
    const message = { ...baseExtensionRangeOptions } as ExtensionRangeOptions;
    message.uninterpretedOption = [];
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseFieldDescriptorProto: object = {
  name: "",
  number: 0,
  label: 1,
  type: 1,
  typeName: "",
  extendee: "",
  defaultValue: "",
  oneofIndex: 0,
  jsonName: "",
  proto3Optional: false,
};

export const FieldDescriptorProto = {
  encode(
    message: FieldDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.number !== 0) {
      writer.uint32(24).int32(message.number);
    }
    if (message.label !== 1) {
      writer.uint32(32).int32(message.label);
    }
    if (message.type !== 1) {
      writer.uint32(40).int32(message.type);
    }
    if (message.typeName !== "") {
      writer.uint32(50).string(message.typeName);
    }
    if (message.extendee !== "") {
      writer.uint32(18).string(message.extendee);
    }
    if (message.defaultValue !== "") {
      writer.uint32(58).string(message.defaultValue);
    }
    if (message.oneofIndex !== 0) {
      writer.uint32(72).int32(message.oneofIndex);
    }
    if (message.jsonName !== "") {
      writer.uint32(82).string(message.jsonName);
    }
    if (message.options !== undefined) {
      FieldOptions.encode(message.options, writer.uint32(66).fork()).ldelim();
    }
    if (message.proto3Optional === true) {
      writer.uint32(136).bool(message.proto3Optional);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): FieldDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFieldDescriptorProto } as FieldDescriptorProto;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 3:
          message.number = reader.int32();
          break;
        case 4:
          message.label = reader.int32() as any;
          break;
        case 5:
          message.type = reader.int32() as any;
          break;
        case 6:
          message.typeName = reader.string();
          break;
        case 2:
          message.extendee = reader.string();
          break;
        case 7:
          message.defaultValue = reader.string();
          break;
        case 9:
          message.oneofIndex = reader.int32();
          break;
        case 10:
          message.jsonName = reader.string();
          break;
        case 8:
          message.options = FieldOptions.decode(reader, reader.uint32());
          break;
        case 17:
          message.proto3Optional = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FieldDescriptorProto {
    const message = { ...baseFieldDescriptorProto } as FieldDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.number !== undefined && object.number !== null) {
      message.number = Number(object.number);
    } else {
      message.number = 0;
    }
    if (object.label !== undefined && object.label !== null) {
      message.label = fieldDescriptorProto_LabelFromJSON(object.label);
    } else {
      message.label = 1;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = fieldDescriptorProto_TypeFromJSON(object.type);
    } else {
      message.type = 1;
    }
    if (object.typeName !== undefined && object.typeName !== null) {
      message.typeName = String(object.typeName);
    } else {
      message.typeName = "";
    }
    if (object.extendee !== undefined && object.extendee !== null) {
      message.extendee = String(object.extendee);
    } else {
      message.extendee = "";
    }
    if (object.defaultValue !== undefined && object.defaultValue !== null) {
      message.defaultValue = String(object.defaultValue);
    } else {
      message.defaultValue = "";
    }
    if (object.oneofIndex !== undefined && object.oneofIndex !== null) {
      message.oneofIndex = Number(object.oneofIndex);
    } else {
      message.oneofIndex = 0;
    }
    if (object.jsonName !== undefined && object.jsonName !== null) {
      message.jsonName = String(object.jsonName);
    } else {
      message.jsonName = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = FieldOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    if (object.proto3Optional !== undefined && object.proto3Optional !== null) {
      message.proto3Optional = Boolean(object.proto3Optional);
    } else {
      message.proto3Optional = false;
    }
    return message;
  },

  toJSON(message: FieldDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.number !== undefined && (obj.number = message.number);
    message.label !== undefined &&
      (obj.label = fieldDescriptorProto_LabelToJSON(message.label));
    message.type !== undefined &&
      (obj.type = fieldDescriptorProto_TypeToJSON(message.type));
    message.typeName !== undefined && (obj.typeName = message.typeName);
    message.extendee !== undefined && (obj.extendee = message.extendee);
    message.defaultValue !== undefined &&
      (obj.defaultValue = message.defaultValue);
    message.oneofIndex !== undefined && (obj.oneofIndex = message.oneofIndex);
    message.jsonName !== undefined && (obj.jsonName = message.jsonName);
    message.options !== undefined &&
      (obj.options = message.options
        ? FieldOptions.toJSON(message.options)
        : undefined);
    message.proto3Optional !== undefined &&
      (obj.proto3Optional = message.proto3Optional);
    return obj;
  },

  fromPartial(object: DeepPartial<FieldDescriptorProto>): FieldDescriptorProto {
    const message = { ...baseFieldDescriptorProto } as FieldDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.number !== undefined && object.number !== null) {
      message.number = object.number;
    } else {
      message.number = 0;
    }
    if (object.label !== undefined && object.label !== null) {
      message.label = object.label;
    } else {
      message.label = 1;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = object.type;
    } else {
      message.type = 1;
    }
    if (object.typeName !== undefined && object.typeName !== null) {
      message.typeName = object.typeName;
    } else {
      message.typeName = "";
    }
    if (object.extendee !== undefined && object.extendee !== null) {
      message.extendee = object.extendee;
    } else {
      message.extendee = "";
    }
    if (object.defaultValue !== undefined && object.defaultValue !== null) {
      message.defaultValue = object.defaultValue;
    } else {
      message.defaultValue = "";
    }
    if (object.oneofIndex !== undefined && object.oneofIndex !== null) {
      message.oneofIndex = object.oneofIndex;
    } else {
      message.oneofIndex = 0;
    }
    if (object.jsonName !== undefined && object.jsonName !== null) {
      message.jsonName = object.jsonName;
    } else {
      message.jsonName = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = FieldOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    if (object.proto3Optional !== undefined && object.proto3Optional !== null) {
      message.proto3Optional = object.proto3Optional;
    } else {
      message.proto3Optional = false;
    }
    return message;
  },
};

const baseOneofDescriptorProto: object = { name: "" };

export const OneofDescriptorProto = {
  encode(
    message: OneofDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.options !== undefined) {
      OneofOptions.encode(message.options, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): OneofDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseOneofDescriptorProto } as OneofDescriptorProto;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.options = OneofOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OneofDescriptorProto {
    const message = { ...baseOneofDescriptorProto } as OneofDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = OneofOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },

  toJSON(message: OneofDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.options !== undefined &&
      (obj.options = message.options
        ? OneofOptions.toJSON(message.options)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<OneofDescriptorProto>): OneofDescriptorProto {
    const message = { ...baseOneofDescriptorProto } as OneofDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = OneofOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },
};

const baseEnumDescriptorProto: object = { name: "", reservedName: "" };

export const EnumDescriptorProto = {
  encode(
    message: EnumDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    for (const v of message.value) {
      EnumValueDescriptorProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.options !== undefined) {
      EnumOptions.encode(message.options, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.reservedRange) {
      EnumDescriptorProto_EnumReservedRange.encode(
        v!,
        writer.uint32(34).fork()
      ).ldelim();
    }
    for (const v of message.reservedName) {
      writer.uint32(42).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EnumDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseEnumDescriptorProto } as EnumDescriptorProto;
    message.value = [];
    message.reservedRange = [];
    message.reservedName = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.value.push(
            EnumValueDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.options = EnumOptions.decode(reader, reader.uint32());
          break;
        case 4:
          message.reservedRange.push(
            EnumDescriptorProto_EnumReservedRange.decode(
              reader,
              reader.uint32()
            )
          );
          break;
        case 5:
          message.reservedName.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EnumDescriptorProto {
    const message = { ...baseEnumDescriptorProto } as EnumDescriptorProto;
    message.value = [];
    message.reservedRange = [];
    message.reservedName = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.value !== undefined && object.value !== null) {
      for (const e of object.value) {
        message.value.push(EnumValueDescriptorProto.fromJSON(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = EnumOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    if (object.reservedRange !== undefined && object.reservedRange !== null) {
      for (const e of object.reservedRange) {
        message.reservedRange.push(
          EnumDescriptorProto_EnumReservedRange.fromJSON(e)
        );
      }
    }
    if (object.reservedName !== undefined && object.reservedName !== null) {
      for (const e of object.reservedName) {
        message.reservedName.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: EnumDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    if (message.value) {
      obj.value = message.value.map((e) =>
        e ? EnumValueDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.value = [];
    }
    message.options !== undefined &&
      (obj.options = message.options
        ? EnumOptions.toJSON(message.options)
        : undefined);
    if (message.reservedRange) {
      obj.reservedRange = message.reservedRange.map((e) =>
        e ? EnumDescriptorProto_EnumReservedRange.toJSON(e) : undefined
      );
    } else {
      obj.reservedRange = [];
    }
    if (message.reservedName) {
      obj.reservedName = message.reservedName.map((e) => e);
    } else {
      obj.reservedName = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<EnumDescriptorProto>): EnumDescriptorProto {
    const message = { ...baseEnumDescriptorProto } as EnumDescriptorProto;
    message.value = [];
    message.reservedRange = [];
    message.reservedName = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.value !== undefined && object.value !== null) {
      for (const e of object.value) {
        message.value.push(EnumValueDescriptorProto.fromPartial(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = EnumOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    if (object.reservedRange !== undefined && object.reservedRange !== null) {
      for (const e of object.reservedRange) {
        message.reservedRange.push(
          EnumDescriptorProto_EnumReservedRange.fromPartial(e)
        );
      }
    }
    if (object.reservedName !== undefined && object.reservedName !== null) {
      for (const e of object.reservedName) {
        message.reservedName.push(e);
      }
    }
    return message;
  },
};

const baseEnumDescriptorProto_EnumReservedRange: object = { start: 0, end: 0 };

export const EnumDescriptorProto_EnumReservedRange = {
  encode(
    message: EnumDescriptorProto_EnumReservedRange,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.start !== 0) {
      writer.uint32(8).int32(message.start);
    }
    if (message.end !== 0) {
      writer.uint32(16).int32(message.end);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): EnumDescriptorProto_EnumReservedRange {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseEnumDescriptorProto_EnumReservedRange,
    } as EnumDescriptorProto_EnumReservedRange;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.start = reader.int32();
          break;
        case 2:
          message.end = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EnumDescriptorProto_EnumReservedRange {
    const message = {
      ...baseEnumDescriptorProto_EnumReservedRange,
    } as EnumDescriptorProto_EnumReservedRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = Number(object.start);
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = Number(object.end);
    } else {
      message.end = 0;
    }
    return message;
  },

  toJSON(message: EnumDescriptorProto_EnumReservedRange): unknown {
    const obj: any = {};
    message.start !== undefined && (obj.start = message.start);
    message.end !== undefined && (obj.end = message.end);
    return obj;
  },

  fromPartial(
    object: DeepPartial<EnumDescriptorProto_EnumReservedRange>
  ): EnumDescriptorProto_EnumReservedRange {
    const message = {
      ...baseEnumDescriptorProto_EnumReservedRange,
    } as EnumDescriptorProto_EnumReservedRange;
    if (object.start !== undefined && object.start !== null) {
      message.start = object.start;
    } else {
      message.start = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = object.end;
    } else {
      message.end = 0;
    }
    return message;
  },
};

const baseEnumValueDescriptorProto: object = { name: "", number: 0 };

export const EnumValueDescriptorProto = {
  encode(
    message: EnumValueDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.number !== 0) {
      writer.uint32(16).int32(message.number);
    }
    if (message.options !== undefined) {
      EnumValueOptions.encode(
        message.options,
        writer.uint32(26).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): EnumValueDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseEnumValueDescriptorProto,
    } as EnumValueDescriptorProto;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.number = reader.int32();
          break;
        case 3:
          message.options = EnumValueOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EnumValueDescriptorProto {
    const message = {
      ...baseEnumValueDescriptorProto,
    } as EnumValueDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.number !== undefined && object.number !== null) {
      message.number = Number(object.number);
    } else {
      message.number = 0;
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = EnumValueOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },

  toJSON(message: EnumValueDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.number !== undefined && (obj.number = message.number);
    message.options !== undefined &&
      (obj.options = message.options
        ? EnumValueOptions.toJSON(message.options)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<EnumValueDescriptorProto>
  ): EnumValueDescriptorProto {
    const message = {
      ...baseEnumValueDescriptorProto,
    } as EnumValueDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.number !== undefined && object.number !== null) {
      message.number = object.number;
    } else {
      message.number = 0;
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = EnumValueOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },
};

const baseServiceDescriptorProto: object = { name: "" };

export const ServiceDescriptorProto = {
  encode(
    message: ServiceDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    for (const v of message.method) {
      MethodDescriptorProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.options !== undefined) {
      ServiceOptions.encode(message.options, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): ServiceDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseServiceDescriptorProto } as ServiceDescriptorProto;
    message.method = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.method.push(
            MethodDescriptorProto.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.options = ServiceOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ServiceDescriptorProto {
    const message = { ...baseServiceDescriptorProto } as ServiceDescriptorProto;
    message.method = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.method !== undefined && object.method !== null) {
      for (const e of object.method) {
        message.method.push(MethodDescriptorProto.fromJSON(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = ServiceOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },

  toJSON(message: ServiceDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    if (message.method) {
      obj.method = message.method.map((e) =>
        e ? MethodDescriptorProto.toJSON(e) : undefined
      );
    } else {
      obj.method = [];
    }
    message.options !== undefined &&
      (obj.options = message.options
        ? ServiceOptions.toJSON(message.options)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<ServiceDescriptorProto>
  ): ServiceDescriptorProto {
    const message = { ...baseServiceDescriptorProto } as ServiceDescriptorProto;
    message.method = [];
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.method !== undefined && object.method !== null) {
      for (const e of object.method) {
        message.method.push(MethodDescriptorProto.fromPartial(e));
      }
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = ServiceOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    return message;
  },
};

const baseMethodDescriptorProto: object = {
  name: "",
  inputType: "",
  outputType: "",
  clientStreaming: false,
  serverStreaming: false,
};

export const MethodDescriptorProto = {
  encode(
    message: MethodDescriptorProto,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.inputType !== "") {
      writer.uint32(18).string(message.inputType);
    }
    if (message.outputType !== "") {
      writer.uint32(26).string(message.outputType);
    }
    if (message.options !== undefined) {
      MethodOptions.encode(message.options, writer.uint32(34).fork()).ldelim();
    }
    if (message.clientStreaming === true) {
      writer.uint32(40).bool(message.clientStreaming);
    }
    if (message.serverStreaming === true) {
      writer.uint32(48).bool(message.serverStreaming);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): MethodDescriptorProto {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMethodDescriptorProto } as MethodDescriptorProto;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.inputType = reader.string();
          break;
        case 3:
          message.outputType = reader.string();
          break;
        case 4:
          message.options = MethodOptions.decode(reader, reader.uint32());
          break;
        case 5:
          message.clientStreaming = reader.bool();
          break;
        case 6:
          message.serverStreaming = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MethodDescriptorProto {
    const message = { ...baseMethodDescriptorProto } as MethodDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.inputType !== undefined && object.inputType !== null) {
      message.inputType = String(object.inputType);
    } else {
      message.inputType = "";
    }
    if (object.outputType !== undefined && object.outputType !== null) {
      message.outputType = String(object.outputType);
    } else {
      message.outputType = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = MethodOptions.fromJSON(object.options);
    } else {
      message.options = undefined;
    }
    if (
      object.clientStreaming !== undefined &&
      object.clientStreaming !== null
    ) {
      message.clientStreaming = Boolean(object.clientStreaming);
    } else {
      message.clientStreaming = false;
    }
    if (
      object.serverStreaming !== undefined &&
      object.serverStreaming !== null
    ) {
      message.serverStreaming = Boolean(object.serverStreaming);
    } else {
      message.serverStreaming = false;
    }
    return message;
  },

  toJSON(message: MethodDescriptorProto): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.inputType !== undefined && (obj.inputType = message.inputType);
    message.outputType !== undefined && (obj.outputType = message.outputType);
    message.options !== undefined &&
      (obj.options = message.options
        ? MethodOptions.toJSON(message.options)
        : undefined);
    message.clientStreaming !== undefined &&
      (obj.clientStreaming = message.clientStreaming);
    message.serverStreaming !== undefined &&
      (obj.serverStreaming = message.serverStreaming);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MethodDescriptorProto>
  ): MethodDescriptorProto {
    const message = { ...baseMethodDescriptorProto } as MethodDescriptorProto;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.inputType !== undefined && object.inputType !== null) {
      message.inputType = object.inputType;
    } else {
      message.inputType = "";
    }
    if (object.outputType !== undefined && object.outputType !== null) {
      message.outputType = object.outputType;
    } else {
      message.outputType = "";
    }
    if (object.options !== undefined && object.options !== null) {
      message.options = MethodOptions.fromPartial(object.options);
    } else {
      message.options = undefined;
    }
    if (
      object.clientStreaming !== undefined &&
      object.clientStreaming !== null
    ) {
      message.clientStreaming = object.clientStreaming;
    } else {
      message.clientStreaming = false;
    }
    if (
      object.serverStreaming !== undefined &&
      object.serverStreaming !== null
    ) {
      message.serverStreaming = object.serverStreaming;
    } else {
      message.serverStreaming = false;
    }
    return message;
  },
};

const baseFileOptions: object = {
  javaPackage: "",
  javaOuterClassname: "",
  javaMultipleFiles: false,
  javaGenerateEqualsAndHash: false,
  javaStringCheckUtf8: false,
  optimizeFor: 1,
  goPackage: "",
  ccGenericServices: false,
  javaGenericServices: false,
  pyGenericServices: false,
  phpGenericServices: false,
  deprecated: false,
  ccEnableArenas: false,
  objcClassPrefix: "",
  csharpNamespace: "",
  swiftPrefix: "",
  phpClassPrefix: "",
  phpNamespace: "",
  phpMetadataNamespace: "",
  rubyPackage: "",
};

export const FileOptions = {
  encode(
    message: FileOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.javaPackage !== "") {
      writer.uint32(10).string(message.javaPackage);
    }
    if (message.javaOuterClassname !== "") {
      writer.uint32(66).string(message.javaOuterClassname);
    }
    if (message.javaMultipleFiles === true) {
      writer.uint32(80).bool(message.javaMultipleFiles);
    }
    if (message.javaGenerateEqualsAndHash === true) {
      writer.uint32(160).bool(message.javaGenerateEqualsAndHash);
    }
    if (message.javaStringCheckUtf8 === true) {
      writer.uint32(216).bool(message.javaStringCheckUtf8);
    }
    if (message.optimizeFor !== 1) {
      writer.uint32(72).int32(message.optimizeFor);
    }
    if (message.goPackage !== "") {
      writer.uint32(90).string(message.goPackage);
    }
    if (message.ccGenericServices === true) {
      writer.uint32(128).bool(message.ccGenericServices);
    }
    if (message.javaGenericServices === true) {
      writer.uint32(136).bool(message.javaGenericServices);
    }
    if (message.pyGenericServices === true) {
      writer.uint32(144).bool(message.pyGenericServices);
    }
    if (message.phpGenericServices === true) {
      writer.uint32(336).bool(message.phpGenericServices);
    }
    if (message.deprecated === true) {
      writer.uint32(184).bool(message.deprecated);
    }
    if (message.ccEnableArenas === true) {
      writer.uint32(248).bool(message.ccEnableArenas);
    }
    if (message.objcClassPrefix !== "") {
      writer.uint32(290).string(message.objcClassPrefix);
    }
    if (message.csharpNamespace !== "") {
      writer.uint32(298).string(message.csharpNamespace);
    }
    if (message.swiftPrefix !== "") {
      writer.uint32(314).string(message.swiftPrefix);
    }
    if (message.phpClassPrefix !== "") {
      writer.uint32(322).string(message.phpClassPrefix);
    }
    if (message.phpNamespace !== "") {
      writer.uint32(330).string(message.phpNamespace);
    }
    if (message.phpMetadataNamespace !== "") {
      writer.uint32(354).string(message.phpMetadataNamespace);
    }
    if (message.rubyPackage !== "") {
      writer.uint32(362).string(message.rubyPackage);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFileOptions } as FileOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.javaPackage = reader.string();
          break;
        case 8:
          message.javaOuterClassname = reader.string();
          break;
        case 10:
          message.javaMultipleFiles = reader.bool();
          break;
        case 20:
          message.javaGenerateEqualsAndHash = reader.bool();
          break;
        case 27:
          message.javaStringCheckUtf8 = reader.bool();
          break;
        case 9:
          message.optimizeFor = reader.int32() as any;
          break;
        case 11:
          message.goPackage = reader.string();
          break;
        case 16:
          message.ccGenericServices = reader.bool();
          break;
        case 17:
          message.javaGenericServices = reader.bool();
          break;
        case 18:
          message.pyGenericServices = reader.bool();
          break;
        case 42:
          message.phpGenericServices = reader.bool();
          break;
        case 23:
          message.deprecated = reader.bool();
          break;
        case 31:
          message.ccEnableArenas = reader.bool();
          break;
        case 36:
          message.objcClassPrefix = reader.string();
          break;
        case 37:
          message.csharpNamespace = reader.string();
          break;
        case 39:
          message.swiftPrefix = reader.string();
          break;
        case 40:
          message.phpClassPrefix = reader.string();
          break;
        case 41:
          message.phpNamespace = reader.string();
          break;
        case 44:
          message.phpMetadataNamespace = reader.string();
          break;
        case 45:
          message.rubyPackage = reader.string();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FileOptions {
    const message = { ...baseFileOptions } as FileOptions;
    message.uninterpretedOption = [];
    if (object.javaPackage !== undefined && object.javaPackage !== null) {
      message.javaPackage = String(object.javaPackage);
    } else {
      message.javaPackage = "";
    }
    if (
      object.javaOuterClassname !== undefined &&
      object.javaOuterClassname !== null
    ) {
      message.javaOuterClassname = String(object.javaOuterClassname);
    } else {
      message.javaOuterClassname = "";
    }
    if (
      object.javaMultipleFiles !== undefined &&
      object.javaMultipleFiles !== null
    ) {
      message.javaMultipleFiles = Boolean(object.javaMultipleFiles);
    } else {
      message.javaMultipleFiles = false;
    }
    if (
      object.javaGenerateEqualsAndHash !== undefined &&
      object.javaGenerateEqualsAndHash !== null
    ) {
      message.javaGenerateEqualsAndHash = Boolean(
        object.javaGenerateEqualsAndHash
      );
    } else {
      message.javaGenerateEqualsAndHash = false;
    }
    if (
      object.javaStringCheckUtf8 !== undefined &&
      object.javaStringCheckUtf8 !== null
    ) {
      message.javaStringCheckUtf8 = Boolean(object.javaStringCheckUtf8);
    } else {
      message.javaStringCheckUtf8 = false;
    }
    if (object.optimizeFor !== undefined && object.optimizeFor !== null) {
      message.optimizeFor = fileOptions_OptimizeModeFromJSON(
        object.optimizeFor
      );
    } else {
      message.optimizeFor = 1;
    }
    if (object.goPackage !== undefined && object.goPackage !== null) {
      message.goPackage = String(object.goPackage);
    } else {
      message.goPackage = "";
    }
    if (
      object.ccGenericServices !== undefined &&
      object.ccGenericServices !== null
    ) {
      message.ccGenericServices = Boolean(object.ccGenericServices);
    } else {
      message.ccGenericServices = false;
    }
    if (
      object.javaGenericServices !== undefined &&
      object.javaGenericServices !== null
    ) {
      message.javaGenericServices = Boolean(object.javaGenericServices);
    } else {
      message.javaGenericServices = false;
    }
    if (
      object.pyGenericServices !== undefined &&
      object.pyGenericServices !== null
    ) {
      message.pyGenericServices = Boolean(object.pyGenericServices);
    } else {
      message.pyGenericServices = false;
    }
    if (
      object.phpGenericServices !== undefined &&
      object.phpGenericServices !== null
    ) {
      message.phpGenericServices = Boolean(object.phpGenericServices);
    } else {
      message.phpGenericServices = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (object.ccEnableArenas !== undefined && object.ccEnableArenas !== null) {
      message.ccEnableArenas = Boolean(object.ccEnableArenas);
    } else {
      message.ccEnableArenas = false;
    }
    if (
      object.objcClassPrefix !== undefined &&
      object.objcClassPrefix !== null
    ) {
      message.objcClassPrefix = String(object.objcClassPrefix);
    } else {
      message.objcClassPrefix = "";
    }
    if (
      object.csharpNamespace !== undefined &&
      object.csharpNamespace !== null
    ) {
      message.csharpNamespace = String(object.csharpNamespace);
    } else {
      message.csharpNamespace = "";
    }
    if (object.swiftPrefix !== undefined && object.swiftPrefix !== null) {
      message.swiftPrefix = String(object.swiftPrefix);
    } else {
      message.swiftPrefix = "";
    }
    if (object.phpClassPrefix !== undefined && object.phpClassPrefix !== null) {
      message.phpClassPrefix = String(object.phpClassPrefix);
    } else {
      message.phpClassPrefix = "";
    }
    if (object.phpNamespace !== undefined && object.phpNamespace !== null) {
      message.phpNamespace = String(object.phpNamespace);
    } else {
      message.phpNamespace = "";
    }
    if (
      object.phpMetadataNamespace !== undefined &&
      object.phpMetadataNamespace !== null
    ) {
      message.phpMetadataNamespace = String(object.phpMetadataNamespace);
    } else {
      message.phpMetadataNamespace = "";
    }
    if (object.rubyPackage !== undefined && object.rubyPackage !== null) {
      message.rubyPackage = String(object.rubyPackage);
    } else {
      message.rubyPackage = "";
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: FileOptions): unknown {
    const obj: any = {};
    message.javaPackage !== undefined &&
      (obj.javaPackage = message.javaPackage);
    message.javaOuterClassname !== undefined &&
      (obj.javaOuterClassname = message.javaOuterClassname);
    message.javaMultipleFiles !== undefined &&
      (obj.javaMultipleFiles = message.javaMultipleFiles);
    message.javaGenerateEqualsAndHash !== undefined &&
      (obj.javaGenerateEqualsAndHash = message.javaGenerateEqualsAndHash);
    message.javaStringCheckUtf8 !== undefined &&
      (obj.javaStringCheckUtf8 = message.javaStringCheckUtf8);
    message.optimizeFor !== undefined &&
      (obj.optimizeFor = fileOptions_OptimizeModeToJSON(message.optimizeFor));
    message.goPackage !== undefined && (obj.goPackage = message.goPackage);
    message.ccGenericServices !== undefined &&
      (obj.ccGenericServices = message.ccGenericServices);
    message.javaGenericServices !== undefined &&
      (obj.javaGenericServices = message.javaGenericServices);
    message.pyGenericServices !== undefined &&
      (obj.pyGenericServices = message.pyGenericServices);
    message.phpGenericServices !== undefined &&
      (obj.phpGenericServices = message.phpGenericServices);
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    message.ccEnableArenas !== undefined &&
      (obj.ccEnableArenas = message.ccEnableArenas);
    message.objcClassPrefix !== undefined &&
      (obj.objcClassPrefix = message.objcClassPrefix);
    message.csharpNamespace !== undefined &&
      (obj.csharpNamespace = message.csharpNamespace);
    message.swiftPrefix !== undefined &&
      (obj.swiftPrefix = message.swiftPrefix);
    message.phpClassPrefix !== undefined &&
      (obj.phpClassPrefix = message.phpClassPrefix);
    message.phpNamespace !== undefined &&
      (obj.phpNamespace = message.phpNamespace);
    message.phpMetadataNamespace !== undefined &&
      (obj.phpMetadataNamespace = message.phpMetadataNamespace);
    message.rubyPackage !== undefined &&
      (obj.rubyPackage = message.rubyPackage);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<FileOptions>): FileOptions {
    const message = { ...baseFileOptions } as FileOptions;
    message.uninterpretedOption = [];
    if (object.javaPackage !== undefined && object.javaPackage !== null) {
      message.javaPackage = object.javaPackage;
    } else {
      message.javaPackage = "";
    }
    if (
      object.javaOuterClassname !== undefined &&
      object.javaOuterClassname !== null
    ) {
      message.javaOuterClassname = object.javaOuterClassname;
    } else {
      message.javaOuterClassname = "";
    }
    if (
      object.javaMultipleFiles !== undefined &&
      object.javaMultipleFiles !== null
    ) {
      message.javaMultipleFiles = object.javaMultipleFiles;
    } else {
      message.javaMultipleFiles = false;
    }
    if (
      object.javaGenerateEqualsAndHash !== undefined &&
      object.javaGenerateEqualsAndHash !== null
    ) {
      message.javaGenerateEqualsAndHash = object.javaGenerateEqualsAndHash;
    } else {
      message.javaGenerateEqualsAndHash = false;
    }
    if (
      object.javaStringCheckUtf8 !== undefined &&
      object.javaStringCheckUtf8 !== null
    ) {
      message.javaStringCheckUtf8 = object.javaStringCheckUtf8;
    } else {
      message.javaStringCheckUtf8 = false;
    }
    if (object.optimizeFor !== undefined && object.optimizeFor !== null) {
      message.optimizeFor = object.optimizeFor;
    } else {
      message.optimizeFor = 1;
    }
    if (object.goPackage !== undefined && object.goPackage !== null) {
      message.goPackage = object.goPackage;
    } else {
      message.goPackage = "";
    }
    if (
      object.ccGenericServices !== undefined &&
      object.ccGenericServices !== null
    ) {
      message.ccGenericServices = object.ccGenericServices;
    } else {
      message.ccGenericServices = false;
    }
    if (
      object.javaGenericServices !== undefined &&
      object.javaGenericServices !== null
    ) {
      message.javaGenericServices = object.javaGenericServices;
    } else {
      message.javaGenericServices = false;
    }
    if (
      object.pyGenericServices !== undefined &&
      object.pyGenericServices !== null
    ) {
      message.pyGenericServices = object.pyGenericServices;
    } else {
      message.pyGenericServices = false;
    }
    if (
      object.phpGenericServices !== undefined &&
      object.phpGenericServices !== null
    ) {
      message.phpGenericServices = object.phpGenericServices;
    } else {
      message.phpGenericServices = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (object.ccEnableArenas !== undefined && object.ccEnableArenas !== null) {
      message.ccEnableArenas = object.ccEnableArenas;
    } else {
      message.ccEnableArenas = false;
    }
    if (
      object.objcClassPrefix !== undefined &&
      object.objcClassPrefix !== null
    ) {
      message.objcClassPrefix = object.objcClassPrefix;
    } else {
      message.objcClassPrefix = "";
    }
    if (
      object.csharpNamespace !== undefined &&
      object.csharpNamespace !== null
    ) {
      message.csharpNamespace = object.csharpNamespace;
    } else {
      message.csharpNamespace = "";
    }
    if (object.swiftPrefix !== undefined && object.swiftPrefix !== null) {
      message.swiftPrefix = object.swiftPrefix;
    } else {
      message.swiftPrefix = "";
    }
    if (object.phpClassPrefix !== undefined && object.phpClassPrefix !== null) {
      message.phpClassPrefix = object.phpClassPrefix;
    } else {
      message.phpClassPrefix = "";
    }
    if (object.phpNamespace !== undefined && object.phpNamespace !== null) {
      message.phpNamespace = object.phpNamespace;
    } else {
      message.phpNamespace = "";
    }
    if (
      object.phpMetadataNamespace !== undefined &&
      object.phpMetadataNamespace !== null
    ) {
      message.phpMetadataNamespace = object.phpMetadataNamespace;
    } else {
      message.phpMetadataNamespace = "";
    }
    if (object.rubyPackage !== undefined && object.rubyPackage !== null) {
      message.rubyPackage = object.rubyPackage;
    } else {
      message.rubyPackage = "";
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseMessageOptions: object = {
  messageSetWireFormat: false,
  noStandardDescriptorAccessor: false,
  deprecated: false,
  mapEntry: false,
};

export const MessageOptions = {
  encode(
    message: MessageOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.messageSetWireFormat === true) {
      writer.uint32(8).bool(message.messageSetWireFormat);
    }
    if (message.noStandardDescriptorAccessor === true) {
      writer.uint32(16).bool(message.noStandardDescriptorAccessor);
    }
    if (message.deprecated === true) {
      writer.uint32(24).bool(message.deprecated);
    }
    if (message.mapEntry === true) {
      writer.uint32(56).bool(message.mapEntry);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MessageOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMessageOptions } as MessageOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.messageSetWireFormat = reader.bool();
          break;
        case 2:
          message.noStandardDescriptorAccessor = reader.bool();
          break;
        case 3:
          message.deprecated = reader.bool();
          break;
        case 7:
          message.mapEntry = reader.bool();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MessageOptions {
    const message = { ...baseMessageOptions } as MessageOptions;
    message.uninterpretedOption = [];
    if (
      object.messageSetWireFormat !== undefined &&
      object.messageSetWireFormat !== null
    ) {
      message.messageSetWireFormat = Boolean(object.messageSetWireFormat);
    } else {
      message.messageSetWireFormat = false;
    }
    if (
      object.noStandardDescriptorAccessor !== undefined &&
      object.noStandardDescriptorAccessor !== null
    ) {
      message.noStandardDescriptorAccessor = Boolean(
        object.noStandardDescriptorAccessor
      );
    } else {
      message.noStandardDescriptorAccessor = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (object.mapEntry !== undefined && object.mapEntry !== null) {
      message.mapEntry = Boolean(object.mapEntry);
    } else {
      message.mapEntry = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: MessageOptions): unknown {
    const obj: any = {};
    message.messageSetWireFormat !== undefined &&
      (obj.messageSetWireFormat = message.messageSetWireFormat);
    message.noStandardDescriptorAccessor !== undefined &&
      (obj.noStandardDescriptorAccessor = message.noStandardDescriptorAccessor);
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    message.mapEntry !== undefined && (obj.mapEntry = message.mapEntry);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<MessageOptions>): MessageOptions {
    const message = { ...baseMessageOptions } as MessageOptions;
    message.uninterpretedOption = [];
    if (
      object.messageSetWireFormat !== undefined &&
      object.messageSetWireFormat !== null
    ) {
      message.messageSetWireFormat = object.messageSetWireFormat;
    } else {
      message.messageSetWireFormat = false;
    }
    if (
      object.noStandardDescriptorAccessor !== undefined &&
      object.noStandardDescriptorAccessor !== null
    ) {
      message.noStandardDescriptorAccessor =
        object.noStandardDescriptorAccessor;
    } else {
      message.noStandardDescriptorAccessor = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (object.mapEntry !== undefined && object.mapEntry !== null) {
      message.mapEntry = object.mapEntry;
    } else {
      message.mapEntry = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseFieldOptions: object = {
  ctype: 0,
  packed: false,
  jstype: 0,
  lazy: false,
  deprecated: false,
  weak: false,
};

export const FieldOptions = {
  encode(
    message: FieldOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.ctype !== 0) {
      writer.uint32(8).int32(message.ctype);
    }
    if (message.packed === true) {
      writer.uint32(16).bool(message.packed);
    }
    if (message.jstype !== 0) {
      writer.uint32(48).int32(message.jstype);
    }
    if (message.lazy === true) {
      writer.uint32(40).bool(message.lazy);
    }
    if (message.deprecated === true) {
      writer.uint32(24).bool(message.deprecated);
    }
    if (message.weak === true) {
      writer.uint32(80).bool(message.weak);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FieldOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFieldOptions } as FieldOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ctype = reader.int32() as any;
          break;
        case 2:
          message.packed = reader.bool();
          break;
        case 6:
          message.jstype = reader.int32() as any;
          break;
        case 5:
          message.lazy = reader.bool();
          break;
        case 3:
          message.deprecated = reader.bool();
          break;
        case 10:
          message.weak = reader.bool();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FieldOptions {
    const message = { ...baseFieldOptions } as FieldOptions;
    message.uninterpretedOption = [];
    if (object.ctype !== undefined && object.ctype !== null) {
      message.ctype = fieldOptions_CTypeFromJSON(object.ctype);
    } else {
      message.ctype = 0;
    }
    if (object.packed !== undefined && object.packed !== null) {
      message.packed = Boolean(object.packed);
    } else {
      message.packed = false;
    }
    if (object.jstype !== undefined && object.jstype !== null) {
      message.jstype = fieldOptions_JSTypeFromJSON(object.jstype);
    } else {
      message.jstype = 0;
    }
    if (object.lazy !== undefined && object.lazy !== null) {
      message.lazy = Boolean(object.lazy);
    } else {
      message.lazy = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (object.weak !== undefined && object.weak !== null) {
      message.weak = Boolean(object.weak);
    } else {
      message.weak = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: FieldOptions): unknown {
    const obj: any = {};
    message.ctype !== undefined &&
      (obj.ctype = fieldOptions_CTypeToJSON(message.ctype));
    message.packed !== undefined && (obj.packed = message.packed);
    message.jstype !== undefined &&
      (obj.jstype = fieldOptions_JSTypeToJSON(message.jstype));
    message.lazy !== undefined && (obj.lazy = message.lazy);
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    message.weak !== undefined && (obj.weak = message.weak);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<FieldOptions>): FieldOptions {
    const message = { ...baseFieldOptions } as FieldOptions;
    message.uninterpretedOption = [];
    if (object.ctype !== undefined && object.ctype !== null) {
      message.ctype = object.ctype;
    } else {
      message.ctype = 0;
    }
    if (object.packed !== undefined && object.packed !== null) {
      message.packed = object.packed;
    } else {
      message.packed = false;
    }
    if (object.jstype !== undefined && object.jstype !== null) {
      message.jstype = object.jstype;
    } else {
      message.jstype = 0;
    }
    if (object.lazy !== undefined && object.lazy !== null) {
      message.lazy = object.lazy;
    } else {
      message.lazy = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (object.weak !== undefined && object.weak !== null) {
      message.weak = object.weak;
    } else {
      message.weak = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseOneofOptions: object = {};

export const OneofOptions = {
  encode(
    message: OneofOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OneofOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseOneofOptions } as OneofOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OneofOptions {
    const message = { ...baseOneofOptions } as OneofOptions;
    message.uninterpretedOption = [];
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: OneofOptions): unknown {
    const obj: any = {};
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<OneofOptions>): OneofOptions {
    const message = { ...baseOneofOptions } as OneofOptions;
    message.uninterpretedOption = [];
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseEnumOptions: object = { allowAlias: false, deprecated: false };

export const EnumOptions = {
  encode(
    message: EnumOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.allowAlias === true) {
      writer.uint32(16).bool(message.allowAlias);
    }
    if (message.deprecated === true) {
      writer.uint32(24).bool(message.deprecated);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EnumOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseEnumOptions } as EnumOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          message.allowAlias = reader.bool();
          break;
        case 3:
          message.deprecated = reader.bool();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EnumOptions {
    const message = { ...baseEnumOptions } as EnumOptions;
    message.uninterpretedOption = [];
    if (object.allowAlias !== undefined && object.allowAlias !== null) {
      message.allowAlias = Boolean(object.allowAlias);
    } else {
      message.allowAlias = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: EnumOptions): unknown {
    const obj: any = {};
    message.allowAlias !== undefined && (obj.allowAlias = message.allowAlias);
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<EnumOptions>): EnumOptions {
    const message = { ...baseEnumOptions } as EnumOptions;
    message.uninterpretedOption = [];
    if (object.allowAlias !== undefined && object.allowAlias !== null) {
      message.allowAlias = object.allowAlias;
    } else {
      message.allowAlias = false;
    }
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseEnumValueOptions: object = { deprecated: false };

export const EnumValueOptions = {
  encode(
    message: EnumValueOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.deprecated === true) {
      writer.uint32(8).bool(message.deprecated);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EnumValueOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseEnumValueOptions } as EnumValueOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.deprecated = reader.bool();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EnumValueOptions {
    const message = { ...baseEnumValueOptions } as EnumValueOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: EnumValueOptions): unknown {
    const obj: any = {};
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<EnumValueOptions>): EnumValueOptions {
    const message = { ...baseEnumValueOptions } as EnumValueOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseServiceOptions: object = { deprecated: false };

export const ServiceOptions = {
  encode(
    message: ServiceOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.deprecated === true) {
      writer.uint32(264).bool(message.deprecated);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ServiceOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseServiceOptions } as ServiceOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 33:
          message.deprecated = reader.bool();
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ServiceOptions {
    const message = { ...baseServiceOptions } as ServiceOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: ServiceOptions): unknown {
    const obj: any = {};
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<ServiceOptions>): ServiceOptions {
    const message = { ...baseServiceOptions } as ServiceOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseMethodOptions: object = { deprecated: false, idempotencyLevel: 0 };

export const MethodOptions = {
  encode(
    message: MethodOptions,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.deprecated === true) {
      writer.uint32(264).bool(message.deprecated);
    }
    if (message.idempotencyLevel !== 0) {
      writer.uint32(272).int32(message.idempotencyLevel);
    }
    for (const v of message.uninterpretedOption) {
      UninterpretedOption.encode(v!, writer.uint32(7994).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MethodOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMethodOptions } as MethodOptions;
    message.uninterpretedOption = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 33:
          message.deprecated = reader.bool();
          break;
        case 34:
          message.idempotencyLevel = reader.int32() as any;
          break;
        case 999:
          message.uninterpretedOption.push(
            UninterpretedOption.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MethodOptions {
    const message = { ...baseMethodOptions } as MethodOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = Boolean(object.deprecated);
    } else {
      message.deprecated = false;
    }
    if (
      object.idempotencyLevel !== undefined &&
      object.idempotencyLevel !== null
    ) {
      message.idempotencyLevel = methodOptions_IdempotencyLevelFromJSON(
        object.idempotencyLevel
      );
    } else {
      message.idempotencyLevel = 0;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: MethodOptions): unknown {
    const obj: any = {};
    message.deprecated !== undefined && (obj.deprecated = message.deprecated);
    message.idempotencyLevel !== undefined &&
      (obj.idempotencyLevel = methodOptions_IdempotencyLevelToJSON(
        message.idempotencyLevel
      ));
    if (message.uninterpretedOption) {
      obj.uninterpretedOption = message.uninterpretedOption.map((e) =>
        e ? UninterpretedOption.toJSON(e) : undefined
      );
    } else {
      obj.uninterpretedOption = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<MethodOptions>): MethodOptions {
    const message = { ...baseMethodOptions } as MethodOptions;
    message.uninterpretedOption = [];
    if (object.deprecated !== undefined && object.deprecated !== null) {
      message.deprecated = object.deprecated;
    } else {
      message.deprecated = false;
    }
    if (
      object.idempotencyLevel !== undefined &&
      object.idempotencyLevel !== null
    ) {
      message.idempotencyLevel = object.idempotencyLevel;
    } else {
      message.idempotencyLevel = 0;
    }
    if (
      object.uninterpretedOption !== undefined &&
      object.uninterpretedOption !== null
    ) {
      for (const e of object.uninterpretedOption) {
        message.uninterpretedOption.push(UninterpretedOption.fromPartial(e));
      }
    }
    return message;
  },
};

const baseUninterpretedOption: object = {
  identifierValue: "",
  positiveIntValue: 0,
  negativeIntValue: 0,
  doubleValue: 0,
  aggregateValue: "",
};

export const UninterpretedOption = {
  encode(
    message: UninterpretedOption,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.name) {
      UninterpretedOption_NamePart.encode(
        v!,
        writer.uint32(18).fork()
      ).ldelim();
    }
    if (message.identifierValue !== "") {
      writer.uint32(26).string(message.identifierValue);
    }
    if (message.positiveIntValue !== 0) {
      writer.uint32(32).uint64(message.positiveIntValue);
    }
    if (message.negativeIntValue !== 0) {
      writer.uint32(40).int64(message.negativeIntValue);
    }
    if (message.doubleValue !== 0) {
      writer.uint32(49).double(message.doubleValue);
    }
    if (message.stringValue.length !== 0) {
      writer.uint32(58).bytes(message.stringValue);
    }
    if (message.aggregateValue !== "") {
      writer.uint32(66).string(message.aggregateValue);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UninterpretedOption {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseUninterpretedOption } as UninterpretedOption;
    message.name = [];
    message.stringValue = new Uint8Array();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          message.name.push(
            UninterpretedOption_NamePart.decode(reader, reader.uint32())
          );
          break;
        case 3:
          message.identifierValue = reader.string();
          break;
        case 4:
          message.positiveIntValue = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.negativeIntValue = longToNumber(reader.int64() as Long);
          break;
        case 6:
          message.doubleValue = reader.double();
          break;
        case 7:
          message.stringValue = reader.bytes();
          break;
        case 8:
          message.aggregateValue = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UninterpretedOption {
    const message = { ...baseUninterpretedOption } as UninterpretedOption;
    message.name = [];
    message.stringValue = new Uint8Array();
    if (object.name !== undefined && object.name !== null) {
      for (const e of object.name) {
        message.name.push(UninterpretedOption_NamePart.fromJSON(e));
      }
    }
    if (
      object.identifierValue !== undefined &&
      object.identifierValue !== null
    ) {
      message.identifierValue = String(object.identifierValue);
    } else {
      message.identifierValue = "";
    }
    if (
      object.positiveIntValue !== undefined &&
      object.positiveIntValue !== null
    ) {
      message.positiveIntValue = Number(object.positiveIntValue);
    } else {
      message.positiveIntValue = 0;
    }
    if (
      object.negativeIntValue !== undefined &&
      object.negativeIntValue !== null
    ) {
      message.negativeIntValue = Number(object.negativeIntValue);
    } else {
      message.negativeIntValue = 0;
    }
    if (object.doubleValue !== undefined && object.doubleValue !== null) {
      message.doubleValue = Number(object.doubleValue);
    } else {
      message.doubleValue = 0;
    }
    if (object.stringValue !== undefined && object.stringValue !== null) {
      message.stringValue = bytesFromBase64(object.stringValue);
    }
    if (object.aggregateValue !== undefined && object.aggregateValue !== null) {
      message.aggregateValue = String(object.aggregateValue);
    } else {
      message.aggregateValue = "";
    }
    return message;
  },

  toJSON(message: UninterpretedOption): unknown {
    const obj: any = {};
    if (message.name) {
      obj.name = message.name.map((e) =>
        e ? UninterpretedOption_NamePart.toJSON(e) : undefined
      );
    } else {
      obj.name = [];
    }
    message.identifierValue !== undefined &&
      (obj.identifierValue = message.identifierValue);
    message.positiveIntValue !== undefined &&
      (obj.positiveIntValue = message.positiveIntValue);
    message.negativeIntValue !== undefined &&
      (obj.negativeIntValue = message.negativeIntValue);
    message.doubleValue !== undefined &&
      (obj.doubleValue = message.doubleValue);
    message.stringValue !== undefined &&
      (obj.stringValue = base64FromBytes(
        message.stringValue !== undefined
          ? message.stringValue
          : new Uint8Array()
      ));
    message.aggregateValue !== undefined &&
      (obj.aggregateValue = message.aggregateValue);
    return obj;
  },

  fromPartial(object: DeepPartial<UninterpretedOption>): UninterpretedOption {
    const message = { ...baseUninterpretedOption } as UninterpretedOption;
    message.name = [];
    if (object.name !== undefined && object.name !== null) {
      for (const e of object.name) {
        message.name.push(UninterpretedOption_NamePart.fromPartial(e));
      }
    }
    if (
      object.identifierValue !== undefined &&
      object.identifierValue !== null
    ) {
      message.identifierValue = object.identifierValue;
    } else {
      message.identifierValue = "";
    }
    if (
      object.positiveIntValue !== undefined &&
      object.positiveIntValue !== null
    ) {
      message.positiveIntValue = object.positiveIntValue;
    } else {
      message.positiveIntValue = 0;
    }
    if (
      object.negativeIntValue !== undefined &&
      object.negativeIntValue !== null
    ) {
      message.negativeIntValue = object.negativeIntValue;
    } else {
      message.negativeIntValue = 0;
    }
    if (object.doubleValue !== undefined && object.doubleValue !== null) {
      message.doubleValue = object.doubleValue;
    } else {
      message.doubleValue = 0;
    }
    if (object.stringValue !== undefined && object.stringValue !== null) {
      message.stringValue = object.stringValue;
    } else {
      message.stringValue = new Uint8Array();
    }
    if (object.aggregateValue !== undefined && object.aggregateValue !== null) {
      message.aggregateValue = object.aggregateValue;
    } else {
      message.aggregateValue = "";
    }
    return message;
  },
};

const baseUninterpretedOption_NamePart: object = {
  namePart: "",
  isExtension: false,
};

export const UninterpretedOption_NamePart = {
  encode(
    message: UninterpretedOption_NamePart,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.namePart !== "") {
      writer.uint32(10).string(message.namePart);
    }
    if (message.isExtension === true) {
      writer.uint32(16).bool(message.isExtension);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): UninterpretedOption_NamePart {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseUninterpretedOption_NamePart,
    } as UninterpretedOption_NamePart;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.namePart = reader.string();
          break;
        case 2:
          message.isExtension = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UninterpretedOption_NamePart {
    const message = {
      ...baseUninterpretedOption_NamePart,
    } as UninterpretedOption_NamePart;
    if (object.namePart !== undefined && object.namePart !== null) {
      message.namePart = String(object.namePart);
    } else {
      message.namePart = "";
    }
    if (object.isExtension !== undefined && object.isExtension !== null) {
      message.isExtension = Boolean(object.isExtension);
    } else {
      message.isExtension = false;
    }
    return message;
  },

  toJSON(message: UninterpretedOption_NamePart): unknown {
    const obj: any = {};
    message.namePart !== undefined && (obj.namePart = message.namePart);
    message.isExtension !== undefined &&
      (obj.isExtension = message.isExtension);
    return obj;
  },

  fromPartial(
    object: DeepPartial<UninterpretedOption_NamePart>
  ): UninterpretedOption_NamePart {
    const message = {
      ...baseUninterpretedOption_NamePart,
    } as UninterpretedOption_NamePart;
    if (object.namePart !== undefined && object.namePart !== null) {
      message.namePart = object.namePart;
    } else {
      message.namePart = "";
    }
    if (object.isExtension !== undefined && object.isExtension !== null) {
      message.isExtension = object.isExtension;
    } else {
      message.isExtension = false;
    }
    return message;
  },
};

const baseSourceCodeInfo: object = {};

export const SourceCodeInfo = {
  encode(
    message: SourceCodeInfo,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.location) {
      SourceCodeInfo_Location.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SourceCodeInfo {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSourceCodeInfo } as SourceCodeInfo;
    message.location = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.location.push(
            SourceCodeInfo_Location.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SourceCodeInfo {
    const message = { ...baseSourceCodeInfo } as SourceCodeInfo;
    message.location = [];
    if (object.location !== undefined && object.location !== null) {
      for (const e of object.location) {
        message.location.push(SourceCodeInfo_Location.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: SourceCodeInfo): unknown {
    const obj: any = {};
    if (message.location) {
      obj.location = message.location.map((e) =>
        e ? SourceCodeInfo_Location.toJSON(e) : undefined
      );
    } else {
      obj.location = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<SourceCodeInfo>): SourceCodeInfo {
    const message = { ...baseSourceCodeInfo } as SourceCodeInfo;
    message.location = [];
    if (object.location !== undefined && object.location !== null) {
      for (const e of object.location) {
        message.location.push(SourceCodeInfo_Location.fromPartial(e));
      }
    }
    return message;
  },
};

const baseSourceCodeInfo_Location: object = {
  path: 0,
  span: 0,
  leadingComments: "",
  trailingComments: "",
  leadingDetachedComments: "",
};

export const SourceCodeInfo_Location = {
  encode(
    message: SourceCodeInfo_Location,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    writer.uint32(10).fork();
    for (const v of message.path) {
      writer.int32(v);
    }
    writer.ldelim();
    writer.uint32(18).fork();
    for (const v of message.span) {
      writer.int32(v);
    }
    writer.ldelim();
    if (message.leadingComments !== "") {
      writer.uint32(26).string(message.leadingComments);
    }
    if (message.trailingComments !== "") {
      writer.uint32(34).string(message.trailingComments);
    }
    for (const v of message.leadingDetachedComments) {
      writer.uint32(50).string(v!);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): SourceCodeInfo_Location {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseSourceCodeInfo_Location,
    } as SourceCodeInfo_Location;
    message.path = [];
    message.span = [];
    message.leadingDetachedComments = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.path.push(reader.int32());
            }
          } else {
            message.path.push(reader.int32());
          }
          break;
        case 2:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.span.push(reader.int32());
            }
          } else {
            message.span.push(reader.int32());
          }
          break;
        case 3:
          message.leadingComments = reader.string();
          break;
        case 4:
          message.trailingComments = reader.string();
          break;
        case 6:
          message.leadingDetachedComments.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SourceCodeInfo_Location {
    const message = {
      ...baseSourceCodeInfo_Location,
    } as SourceCodeInfo_Location;
    message.path = [];
    message.span = [];
    message.leadingDetachedComments = [];
    if (object.path !== undefined && object.path !== null) {
      for (const e of object.path) {
        message.path.push(Number(e));
      }
    }
    if (object.span !== undefined && object.span !== null) {
      for (const e of object.span) {
        message.span.push(Number(e));
      }
    }
    if (
      object.leadingComments !== undefined &&
      object.leadingComments !== null
    ) {
      message.leadingComments = String(object.leadingComments);
    } else {
      message.leadingComments = "";
    }
    if (
      object.trailingComments !== undefined &&
      object.trailingComments !== null
    ) {
      message.trailingComments = String(object.trailingComments);
    } else {
      message.trailingComments = "";
    }
    if (
      object.leadingDetachedComments !== undefined &&
      object.leadingDetachedComments !== null
    ) {
      for (const e of object.leadingDetachedComments) {
        message.leadingDetachedComments.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: SourceCodeInfo_Location): unknown {
    const obj: any = {};
    if (message.path) {
      obj.path = message.path.map((e) => e);
    } else {
      obj.path = [];
    }
    if (message.span) {
      obj.span = message.span.map((e) => e);
    } else {
      obj.span = [];
    }
    message.leadingComments !== undefined &&
      (obj.leadingComments = message.leadingComments);
    message.trailingComments !== undefined &&
      (obj.trailingComments = message.trailingComments);
    if (message.leadingDetachedComments) {
      obj.leadingDetachedComments = message.leadingDetachedComments.map(
        (e) => e
      );
    } else {
      obj.leadingDetachedComments = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<SourceCodeInfo_Location>
  ): SourceCodeInfo_Location {
    const message = {
      ...baseSourceCodeInfo_Location,
    } as SourceCodeInfo_Location;
    message.path = [];
    message.span = [];
    message.leadingDetachedComments = [];
    if (object.path !== undefined && object.path !== null) {
      for (const e of object.path) {
        message.path.push(e);
      }
    }
    if (object.span !== undefined && object.span !== null) {
      for (const e of object.span) {
        message.span.push(e);
      }
    }
    if (
      object.leadingComments !== undefined &&
      object.leadingComments !== null
    ) {
      message.leadingComments = object.leadingComments;
    } else {
      message.leadingComments = "";
    }
    if (
      object.trailingComments !== undefined &&
      object.trailingComments !== null
    ) {
      message.trailingComments = object.trailingComments;
    } else {
      message.trailingComments = "";
    }
    if (
      object.leadingDetachedComments !== undefined &&
      object.leadingDetachedComments !== null
    ) {
      for (const e of object.leadingDetachedComments) {
        message.leadingDetachedComments.push(e);
      }
    }
    return message;
  },
};

const baseGeneratedCodeInfo: object = {};

export const GeneratedCodeInfo = {
  encode(
    message: GeneratedCodeInfo,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.annotation) {
      GeneratedCodeInfo_Annotation.encode(
        v!,
        writer.uint32(10).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GeneratedCodeInfo {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGeneratedCodeInfo } as GeneratedCodeInfo;
    message.annotation = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.annotation.push(
            GeneratedCodeInfo_Annotation.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GeneratedCodeInfo {
    const message = { ...baseGeneratedCodeInfo } as GeneratedCodeInfo;
    message.annotation = [];
    if (object.annotation !== undefined && object.annotation !== null) {
      for (const e of object.annotation) {
        message.annotation.push(GeneratedCodeInfo_Annotation.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GeneratedCodeInfo): unknown {
    const obj: any = {};
    if (message.annotation) {
      obj.annotation = message.annotation.map((e) =>
        e ? GeneratedCodeInfo_Annotation.toJSON(e) : undefined
      );
    } else {
      obj.annotation = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GeneratedCodeInfo>): GeneratedCodeInfo {
    const message = { ...baseGeneratedCodeInfo } as GeneratedCodeInfo;
    message.annotation = [];
    if (object.annotation !== undefined && object.annotation !== null) {
      for (const e of object.annotation) {
        message.annotation.push(GeneratedCodeInfo_Annotation.fromPartial(e));
      }
    }
    return message;
  },
};

const baseGeneratedCodeInfo_Annotation: object = {
  path: 0,
  sourceFile: "",
  begin: 0,
  end: 0,
};

export const GeneratedCodeInfo_Annotation = {
  encode(
    message: GeneratedCodeInfo_Annotation,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    writer.uint32(10).fork();
    for (const v of message.path) {
      writer.int32(v);
    }
    writer.ldelim();
    if (message.sourceFile !== "") {
      writer.uint32(18).string(message.sourceFile);
    }
    if (message.begin !== 0) {
      writer.uint32(24).int32(message.begin);
    }
    if (message.end !== 0) {
      writer.uint32(32).int32(message.end);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): GeneratedCodeInfo_Annotation {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGeneratedCodeInfo_Annotation,
    } as GeneratedCodeInfo_Annotation;
    message.path = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.path.push(reader.int32());
            }
          } else {
            message.path.push(reader.int32());
          }
          break;
        case 2:
          message.sourceFile = reader.string();
          break;
        case 3:
          message.begin = reader.int32();
          break;
        case 4:
          message.end = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GeneratedCodeInfo_Annotation {
    const message = {
      ...baseGeneratedCodeInfo_Annotation,
    } as GeneratedCodeInfo_Annotation;
    message.path = [];
    if (object.path !== undefined && object.path !== null) {
      for (const e of object.path) {
        message.path.push(Number(e));
      }
    }
    if (object.sourceFile !== undefined && object.sourceFile !== null) {
      message.sourceFile = String(object.sourceFile);
    } else {
      message.sourceFile = "";
    }
    if (object.begin !== undefined && object.begin !== null) {
      message.begin = Number(object.begin);
    } else {
      message.begin = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = Number(object.end);
    } else {
      message.end = 0;
    }
    return message;
  },

  toJSON(message: GeneratedCodeInfo_Annotation): unknown {
    const obj: any = {};
    if (message.path) {
      obj.path = message.path.map((e) => e);
    } else {
      obj.path = [];
    }
    message.sourceFile !== undefined && (obj.sourceFile = message.sourceFile);
    message.begin !== undefined && (obj.begin = message.begin);
    message.end !== undefined && (obj.end = message.end);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GeneratedCodeInfo_Annotation>
  ): GeneratedCodeInfo_Annotation {
    const message = {
      ...baseGeneratedCodeInfo_Annotation,
    } as GeneratedCodeInfo_Annotation;
    message.path = [];
    if (object.path !== undefined && object.path !== null) {
      for (const e of object.path) {
        message.path.push(e);
      }
    }
    if (object.sourceFile !== undefined && object.sourceFile !== null) {
      message.sourceFile = object.sourceFile;
    } else {
      message.sourceFile = "";
    }
    if (object.begin !== undefined && object.begin !== null) {
      message.begin = object.begin;
    } else {
      message.begin = 0;
    }
    if (object.end !== undefined && object.end !== null) {
      message.end = object.end;
    } else {
      message.end = 0;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
declare var global: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (const byte of arr) {
    bin.push(String.fromCharCode(byte));
  }
  return btoa(bin.join(""));
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

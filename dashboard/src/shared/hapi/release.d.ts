// tslint:disable
import * as $protobuf from "protobufjs";

/** Namespace hapi. */
export namespace hapi {
  /** Namespace release. */
  namespace release {
    /** Properties of a Release. */
    interface IRelease {
      /** Release name */
      name?: string | null;

      /** Release info */
      info?: hapi.release.IInfo | null;

      /** Release chart */
      chart?: hapi.chart.IChart | null;

      /** Release config */
      config?: hapi.chart.IConfig | null;

      /** Release manifest */
      manifest?: string | null;

      /** Release hooks */
      hooks?: hapi.release.IHook[] | null;

      /** Release version */
      version?: number | null;

      /** Release namespace */
      namespace?: string | null;
    }

    /** Represents a Release. */
    class Release implements IRelease {
      /**
       * Constructs a new Release.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.IRelease);

      /** Release name. */
      public name: string;

      /** Release info. */
      public info?: hapi.release.IInfo | null;

      /** Release chart. */
      public chart?: hapi.chart.IChart | null;

      /** Release config. */
      public config?: hapi.chart.IConfig | null;

      /** Release manifest. */
      public manifest: string;

      /** Release hooks. */
      public hooks: hapi.release.IHook[];

      /** Release version. */
      public version: number;

      /** Release namespace. */
      public namespace: string;

      /**
       * Creates a new Release instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Release instance
       */
      public static create(properties?: hapi.release.IRelease): hapi.release.Release;

      /**
       * Encodes the specified Release message. Does not implicitly {@link hapi.release.Release.verify|verify} messages.
       * @param message Release message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.IRelease,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Release message, length delimited. Does not implicitly {@link hapi.release.Release.verify|verify} messages.
       * @param message Release message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.IRelease,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Release message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Release
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.Release;

      /**
       * Decodes a Release message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Release
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.Release;

      /**
       * Verifies a Release message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Release message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Release
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.Release;

      /**
       * Creates a plain object from a Release message. Also converts values to other types if specified.
       * @param message Release
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.Release,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Release to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Hook. */
    interface IHook {
      /** Hook name */
      name?: string | null;

      /** Hook kind */
      kind?: string | null;

      /** Hook path */
      path?: string | null;

      /** Hook manifest */
      manifest?: string | null;

      /** Hook events */
      events?: hapi.release.Hook.Event[] | null;

      /** Hook lastRun */
      lastRun?: google.protobuf.ITimestamp | null;

      /** Hook weight */
      weight?: number | null;

      /** Hook deletePolicies */
      deletePolicies?: hapi.release.Hook.DeletePolicy[] | null;
    }

    /** Represents a Hook. */
    class Hook implements IHook {
      /**
       * Constructs a new Hook.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.IHook);

      /** Hook name. */
      public name: string;

      /** Hook kind. */
      public kind: string;

      /** Hook path. */
      public path: string;

      /** Hook manifest. */
      public manifest: string;

      /** Hook events. */
      public events: hapi.release.Hook.Event[];

      /** Hook lastRun. */
      public lastRun?: google.protobuf.ITimestamp | null;

      /** Hook weight. */
      public weight: number;

      /** Hook deletePolicies. */
      public deletePolicies: hapi.release.Hook.DeletePolicy[];

      /**
       * Creates a new Hook instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Hook instance
       */
      public static create(properties?: hapi.release.IHook): hapi.release.Hook;

      /**
       * Encodes the specified Hook message. Does not implicitly {@link hapi.release.Hook.verify|verify} messages.
       * @param message Hook message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.IHook,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Hook message, length delimited. Does not implicitly {@link hapi.release.Hook.verify|verify} messages.
       * @param message Hook message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.IHook,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Hook message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Hook
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.Hook;

      /**
       * Decodes a Hook message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Hook
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.Hook;

      /**
       * Verifies a Hook message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Hook message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Hook
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.Hook;

      /**
       * Creates a plain object from a Hook message. Also converts values to other types if specified.
       * @param message Hook
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.Hook,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Hook to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    namespace Hook {
      /** Event enum. */
      enum Event {
        UNKNOWN = 0,
        PRE_INSTALL = 1,
        POST_INSTALL = 2,
        PRE_DELETE = 3,
        POST_DELETE = 4,
        PRE_UPGRADE = 5,
        POST_UPGRADE = 6,
        PRE_ROLLBACK = 7,
        POST_ROLLBACK = 8,
        RELEASE_TEST_SUCCESS = 9,
        RELEASE_TEST_FAILURE = 10,
      }

      /** DeletePolicy enum. */
      enum DeletePolicy {
        SUCCEEDED = 0,
        FAILED = 1,
      }
    }

    /** Properties of an Info. */
    interface IInfo {
      /** Info status */
      status?: hapi.release.IStatus | null;

      /** Info firstDeployed */
      firstDeployed?: google.protobuf.ITimestamp | null;

      /** Info lastDeployed */
      lastDeployed?: google.protobuf.ITimestamp | null;

      /** Info deleted */
      deleted?: google.protobuf.ITimestamp | null;

      /** Info Description */
      Description?: string | null;
    }

    /** Represents an Info. */
    class Info implements IInfo {
      /**
       * Constructs a new Info.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.IInfo);

      /** Info status. */
      public status?: hapi.release.IStatus | null;

      /** Info firstDeployed. */
      public firstDeployed?: google.protobuf.ITimestamp | null;

      /** Info lastDeployed. */
      public lastDeployed?: google.protobuf.ITimestamp | null;

      /** Info deleted. */
      public deleted?: google.protobuf.ITimestamp | null;

      /** Info Description. */
      public Description: string;

      /**
       * Creates a new Info instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Info instance
       */
      public static create(properties?: hapi.release.IInfo): hapi.release.Info;

      /**
       * Encodes the specified Info message. Does not implicitly {@link hapi.release.Info.verify|verify} messages.
       * @param message Info message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.IInfo,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Info message, length delimited. Does not implicitly {@link hapi.release.Info.verify|verify} messages.
       * @param message Info message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.IInfo,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes an Info message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Info
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.Info;

      /**
       * Decodes an Info message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Info
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.Info;

      /**
       * Verifies an Info message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates an Info message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Info
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.Info;

      /**
       * Creates a plain object from an Info message. Also converts values to other types if specified.
       * @param message Info
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.Info,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Info to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Status. */
    interface IStatus {
      /** Status code */
      code?: hapi.release.Status.Code | null;

      /** Status resources */
      resources?: string | null;

      /** Status notes */
      notes?: string | null;

      /** Status lastTestSuiteRun */
      lastTestSuiteRun?: hapi.release.ITestSuite | null;
    }

    /** Represents a Status. */
    class Status implements IStatus {
      /**
       * Constructs a new Status.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.IStatus);

      /** Status code. */
      public code: hapi.release.Status.Code;

      /** Status resources. */
      public resources: string;

      /** Status notes. */
      public notes: string;

      /** Status lastTestSuiteRun. */
      public lastTestSuiteRun?: hapi.release.ITestSuite | null;

      /**
       * Creates a new Status instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Status instance
       */
      public static create(properties?: hapi.release.IStatus): hapi.release.Status;

      /**
       * Encodes the specified Status message. Does not implicitly {@link hapi.release.Status.verify|verify} messages.
       * @param message Status message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.IStatus,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Status message, length delimited. Does not implicitly {@link hapi.release.Status.verify|verify} messages.
       * @param message Status message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.IStatus,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Status message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Status
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.Status;

      /**
       * Decodes a Status message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Status
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.Status;

      /**
       * Verifies a Status message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Status message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Status
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.Status;

      /**
       * Creates a plain object from a Status message. Also converts values to other types if specified.
       * @param message Status
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.Status,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Status to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    namespace Status {
      /** Code enum. */
      enum Code {
        UNKNOWN = 0,
        DEPLOYED = 1,
        DELETED = 2,
        SUPERSEDED = 3,
        FAILED = 4,
        DELETING = 5,
        PENDING_INSTALL = 6,
        PENDING_UPGRADE = 7,
        PENDING_ROLLBACK = 8,
      }
    }

    /** Properties of a TestSuite. */
    interface ITestSuite {
      /** TestSuite startedAt */
      startedAt?: google.protobuf.ITimestamp | null;

      /** TestSuite completedAt */
      completedAt?: google.protobuf.ITimestamp | null;

      /** TestSuite results */
      results?: hapi.release.ITestRun[] | null;
    }

    /** Represents a TestSuite. */
    class TestSuite implements ITestSuite {
      /**
       * Constructs a new TestSuite.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.ITestSuite);

      /** TestSuite startedAt. */
      public startedAt?: google.protobuf.ITimestamp | null;

      /** TestSuite completedAt. */
      public completedAt?: google.protobuf.ITimestamp | null;

      /** TestSuite results. */
      public results: hapi.release.ITestRun[];

      /**
       * Creates a new TestSuite instance using the specified properties.
       * @param [properties] Properties to set
       * @returns TestSuite instance
       */
      public static create(properties?: hapi.release.ITestSuite): hapi.release.TestSuite;

      /**
       * Encodes the specified TestSuite message. Does not implicitly {@link hapi.release.TestSuite.verify|verify} messages.
       * @param message TestSuite message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.ITestSuite,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified TestSuite message, length delimited. Does not implicitly {@link hapi.release.TestSuite.verify|verify} messages.
       * @param message TestSuite message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.ITestSuite,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a TestSuite message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns TestSuite
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.TestSuite;

      /**
       * Decodes a TestSuite message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns TestSuite
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.TestSuite;

      /**
       * Verifies a TestSuite message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a TestSuite message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns TestSuite
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.TestSuite;

      /**
       * Creates a plain object from a TestSuite message. Also converts values to other types if specified.
       * @param message TestSuite
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.TestSuite,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this TestSuite to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a TestRun. */
    interface ITestRun {
      /** TestRun name */
      name?: string | null;

      /** TestRun status */
      status?: hapi.release.TestRun.Status | null;

      /** TestRun info */
      info?: string | null;

      /** TestRun startedAt */
      startedAt?: google.protobuf.ITimestamp | null;

      /** TestRun completedAt */
      completedAt?: google.protobuf.ITimestamp | null;
    }

    /** Represents a TestRun. */
    class TestRun implements ITestRun {
      /**
       * Constructs a new TestRun.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.release.ITestRun);

      /** TestRun name. */
      public name: string;

      /** TestRun status. */
      public status: hapi.release.TestRun.Status;

      /** TestRun info. */
      public info: string;

      /** TestRun startedAt. */
      public startedAt?: google.protobuf.ITimestamp | null;

      /** TestRun completedAt. */
      public completedAt?: google.protobuf.ITimestamp | null;

      /**
       * Creates a new TestRun instance using the specified properties.
       * @param [properties] Properties to set
       * @returns TestRun instance
       */
      public static create(properties?: hapi.release.ITestRun): hapi.release.TestRun;

      /**
       * Encodes the specified TestRun message. Does not implicitly {@link hapi.release.TestRun.verify|verify} messages.
       * @param message TestRun message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.release.ITestRun,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified TestRun message, length delimited. Does not implicitly {@link hapi.release.TestRun.verify|verify} messages.
       * @param message TestRun message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.release.ITestRun,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a TestRun message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns TestRun
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.release.TestRun;

      /**
       * Decodes a TestRun message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns TestRun
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.release.TestRun;

      /**
       * Verifies a TestRun message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a TestRun message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns TestRun
       */
      public static fromObject(object: { [k: string]: any }): hapi.release.TestRun;

      /**
       * Creates a plain object from a TestRun message. Also converts values to other types if specified.
       * @param message TestRun
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.release.TestRun,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this TestRun to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    namespace TestRun {
      /** Status enum. */
      enum Status {
        UNKNOWN = 0,
        SUCCESS = 1,
        FAILURE = 2,
        RUNNING = 3,
      }
    }
  }

  /** Namespace chart. */
  namespace chart {
    /** Properties of a Config. */
    interface IConfig {
      /** Config raw */
      raw?: string | null;

      /** Config values */
      values?: { [k: string]: hapi.chart.IValue } | null;
    }

    /** Represents a Config. */
    class Config implements IConfig {
      /**
       * Constructs a new Config.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.IConfig);

      /** Config raw. */
      public raw: string;

      /** Config values. */
      public values: { [k: string]: hapi.chart.IValue };

      /**
       * Creates a new Config instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Config instance
       */
      public static create(properties?: hapi.chart.IConfig): hapi.chart.Config;

      /**
       * Encodes the specified Config message. Does not implicitly {@link hapi.chart.Config.verify|verify} messages.
       * @param message Config message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.chart.IConfig,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Config message, length delimited. Does not implicitly {@link hapi.chart.Config.verify|verify} messages.
       * @param message Config message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.IConfig,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Config message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Config
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Config;

      /**
       * Decodes a Config message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Config
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Config;

      /**
       * Verifies a Config message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Config message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Config
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Config;

      /**
       * Creates a plain object from a Config message. Also converts values to other types if specified.
       * @param message Config
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Config,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Config to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Value. */
    interface IValue {
      /** Value value */
      value?: string | null;
    }

    /** Represents a Value. */
    class Value implements IValue {
      /**
       * Constructs a new Value.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.IValue);

      /** Value value. */
      public value: string;

      /**
       * Creates a new Value instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Value instance
       */
      public static create(properties?: hapi.chart.IValue): hapi.chart.Value;

      /**
       * Encodes the specified Value message. Does not implicitly {@link hapi.chart.Value.verify|verify} messages.
       * @param message Value message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(message: hapi.chart.IValue, writer?: $protobuf.Writer): $protobuf.Writer;

      /**
       * Encodes the specified Value message, length delimited. Does not implicitly {@link hapi.chart.Value.verify|verify} messages.
       * @param message Value message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.IValue,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Value message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Value
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Value;

      /**
       * Decodes a Value message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Value
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Value;

      /**
       * Verifies a Value message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Value message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Value
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Value;

      /**
       * Creates a plain object from a Value message. Also converts values to other types if specified.
       * @param message Value
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Value,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Value to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Chart. */
    interface IChart {
      /** Chart metadata */
      metadata?: hapi.chart.IMetadata | null;

      /** Chart templates */
      templates?: hapi.chart.ITemplate[] | null;

      /** Chart dependencies */
      dependencies?: hapi.chart.IChart[] | null;

      /** Chart values */
      values?: hapi.chart.IConfig | null;

      /** Chart files */
      files?: google.protobuf.IAny[] | null;
    }

    /** Represents a Chart. */
    class Chart implements IChart {
      /**
       * Constructs a new Chart.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.IChart);

      /** Chart metadata. */
      public metadata?: hapi.chart.IMetadata | null;

      /** Chart templates. */
      public templates: hapi.chart.ITemplate[];

      /** Chart dependencies. */
      public dependencies: hapi.chart.IChart[];

      /** Chart values. */
      public values?: hapi.chart.IConfig | null;

      /** Chart files. */
      public files: google.protobuf.IAny[];

      /**
       * Creates a new Chart instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Chart instance
       */
      public static create(properties?: hapi.chart.IChart): hapi.chart.Chart;

      /**
       * Encodes the specified Chart message. Does not implicitly {@link hapi.chart.Chart.verify|verify} messages.
       * @param message Chart message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(message: hapi.chart.IChart, writer?: $protobuf.Writer): $protobuf.Writer;

      /**
       * Encodes the specified Chart message, length delimited. Does not implicitly {@link hapi.chart.Chart.verify|verify} messages.
       * @param message Chart message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.IChart,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Chart message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Chart
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Chart;

      /**
       * Decodes a Chart message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Chart
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Chart;

      /**
       * Verifies a Chart message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Chart message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Chart
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Chart;

      /**
       * Creates a plain object from a Chart message. Also converts values to other types if specified.
       * @param message Chart
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Chart,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Chart to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Maintainer. */
    interface IMaintainer {
      /** Maintainer name */
      name?: string | null;

      /** Maintainer email */
      email?: string | null;

      /** Maintainer url */
      url?: string | null;
    }

    /** Represents a Maintainer. */
    class Maintainer implements IMaintainer {
      /**
       * Constructs a new Maintainer.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.IMaintainer);

      /** Maintainer name. */
      public name: string;

      /** Maintainer email. */
      public email: string;

      /** Maintainer url. */
      public url: string;

      /**
       * Creates a new Maintainer instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Maintainer instance
       */
      public static create(properties?: hapi.chart.IMaintainer): hapi.chart.Maintainer;

      /**
       * Encodes the specified Maintainer message. Does not implicitly {@link hapi.chart.Maintainer.verify|verify} messages.
       * @param message Maintainer message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.chart.IMaintainer,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Maintainer message, length delimited. Does not implicitly {@link hapi.chart.Maintainer.verify|verify} messages.
       * @param message Maintainer message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.IMaintainer,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Maintainer message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Maintainer
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Maintainer;

      /**
       * Decodes a Maintainer message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Maintainer
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Maintainer;

      /**
       * Verifies a Maintainer message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Maintainer message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Maintainer
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Maintainer;

      /**
       * Creates a plain object from a Maintainer message. Also converts values to other types if specified.
       * @param message Maintainer
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Maintainer,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Maintainer to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of a Metadata. */
    interface IMetadata {
      /** Metadata name */
      name?: string | null;

      /** Metadata home */
      home?: string | null;

      /** Metadata sources */
      sources?: string[] | null;

      /** Metadata version */
      version?: string | null;

      /** Metadata description */
      description?: string | null;

      /** Metadata keywords */
      keywords?: string[] | null;

      /** Metadata maintainers */
      maintainers?: hapi.chart.IMaintainer[] | null;

      /** Metadata engine */
      engine?: string | null;

      /** Metadata icon */
      icon?: string | null;

      /** Metadata apiVersion */
      apiVersion?: string | null;

      /** Metadata condition */
      condition?: string | null;

      /** Metadata tags */
      tags?: string | null;

      /** Metadata appVersion */
      appVersion?: string | null;

      /** Metadata deprecated */
      deprecated?: boolean | null;

      /** Metadata tillerVersion */
      tillerVersion?: string | null;

      /** Metadata annotations */
      annotations?: { [k: string]: string } | null;

      /** Metadata kubeVersion */
      kubeVersion?: string | null;
    }

    /** Represents a Metadata. */
    class Metadata implements IMetadata {
      /**
       * Constructs a new Metadata.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.IMetadata);

      /** Metadata name. */
      public name: string;

      /** Metadata home. */
      public home: string;

      /** Metadata sources. */
      public sources: string[];

      /** Metadata version. */
      public version: string;

      /** Metadata description. */
      public description: string;

      /** Metadata keywords. */
      public keywords: string[];

      /** Metadata maintainers. */
      public maintainers: hapi.chart.IMaintainer[];

      /** Metadata engine. */
      public engine: string;

      /** Metadata icon. */
      public icon: string;

      /** Metadata apiVersion. */
      public apiVersion: string;

      /** Metadata condition. */
      public condition: string;

      /** Metadata tags. */
      public tags: string;

      /** Metadata appVersion. */
      public appVersion: string;

      /** Metadata deprecated. */
      public deprecated: boolean;

      /** Metadata tillerVersion. */
      public tillerVersion: string;

      /** Metadata annotations. */
      public annotations: { [k: string]: string };

      /** Metadata kubeVersion. */
      public kubeVersion: string;

      /**
       * Creates a new Metadata instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Metadata instance
       */
      public static create(properties?: hapi.chart.IMetadata): hapi.chart.Metadata;

      /**
       * Encodes the specified Metadata message. Does not implicitly {@link hapi.chart.Metadata.verify|verify} messages.
       * @param message Metadata message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.chart.IMetadata,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Metadata message, length delimited. Does not implicitly {@link hapi.chart.Metadata.verify|verify} messages.
       * @param message Metadata message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.IMetadata,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Metadata message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Metadata
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Metadata;

      /**
       * Decodes a Metadata message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Metadata
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Metadata;

      /**
       * Verifies a Metadata message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Metadata message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Metadata
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Metadata;

      /**
       * Creates a plain object from a Metadata message. Also converts values to other types if specified.
       * @param message Metadata
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Metadata,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Metadata to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    namespace Metadata {
      /** Engine enum. */
      enum Engine {
        UNKNOWN = 0,
        GOTPL = 1,
      }
    }

    /** Properties of a Template. */
    interface ITemplate {
      /** Template name */
      name?: string | null;

      /** Template data */
      data?: Uint8Array | null;
    }

    /** Represents a Template. */
    class Template implements ITemplate {
      /**
       * Constructs a new Template.
       * @param [properties] Properties to set
       */
      constructor(properties?: hapi.chart.ITemplate);

      /** Template name. */
      public name: string;

      /** Template data. */
      public data: Uint8Array;

      /**
       * Creates a new Template instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Template instance
       */
      public static create(properties?: hapi.chart.ITemplate): hapi.chart.Template;

      /**
       * Encodes the specified Template message. Does not implicitly {@link hapi.chart.Template.verify|verify} messages.
       * @param message Template message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: hapi.chart.ITemplate,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Template message, length delimited. Does not implicitly {@link hapi.chart.Template.verify|verify} messages.
       * @param message Template message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: hapi.chart.ITemplate,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Template message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Template
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): hapi.chart.Template;

      /**
       * Decodes a Template message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Template
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): hapi.chart.Template;

      /**
       * Verifies a Template message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Template message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Template
       */
      public static fromObject(object: { [k: string]: any }): hapi.chart.Template;

      /**
       * Creates a plain object from a Template message. Also converts values to other types if specified.
       * @param message Template
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: hapi.chart.Template,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Template to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }
  }
}

/** Namespace google. */
export namespace google {
  /** Namespace protobuf. */
  namespace protobuf {
    /** Properties of a Timestamp. */
    interface ITimestamp {
      /** Timestamp seconds */
      seconds?: number | Long | null;

      /** Timestamp nanos */
      nanos?: number | null;
    }

    /** Represents a Timestamp. */
    class Timestamp implements ITimestamp {
      /**
       * Constructs a new Timestamp.
       * @param [properties] Properties to set
       */
      constructor(properties?: google.protobuf.ITimestamp);

      /** Timestamp seconds. */
      public seconds: number | Long;

      /** Timestamp nanos. */
      public nanos: number;

      /**
       * Creates a new Timestamp instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Timestamp instance
       */
      public static create(properties?: google.protobuf.ITimestamp): google.protobuf.Timestamp;

      /**
       * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
       * @param message Timestamp message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: google.protobuf.ITimestamp,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Timestamp message, length delimited. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
       * @param message Timestamp message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: google.protobuf.ITimestamp,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes a Timestamp message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Timestamp
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): google.protobuf.Timestamp;

      /**
       * Decodes a Timestamp message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Timestamp
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(
        reader: $protobuf.Reader | Uint8Array,
      ): google.protobuf.Timestamp;

      /**
       * Verifies a Timestamp message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates a Timestamp message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Timestamp
       */
      public static fromObject(object: { [k: string]: any }): google.protobuf.Timestamp;

      /**
       * Creates a plain object from a Timestamp message. Also converts values to other types if specified.
       * @param message Timestamp
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: google.protobuf.Timestamp,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Timestamp to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }

    /** Properties of an Any. */
    interface IAny {
      /** Any type_url */
      type_url?: string | null;

      /** Any value */
      value?: Uint8Array | null;
    }

    /** Represents an Any. */
    class Any implements IAny {
      /**
       * Constructs a new Any.
       * @param [properties] Properties to set
       */
      constructor(properties?: google.protobuf.IAny);

      /** Any type_url. */
      public type_url: string;

      /** Any value. */
      public value: Uint8Array;

      /**
       * Creates a new Any instance using the specified properties.
       * @param [properties] Properties to set
       * @returns Any instance
       */
      public static create(properties?: google.protobuf.IAny): google.protobuf.Any;

      /**
       * Encodes the specified Any message. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
       * @param message Any message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encode(
        message: google.protobuf.IAny,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Encodes the specified Any message, length delimited. Does not implicitly {@link google.protobuf.Any.verify|verify} messages.
       * @param message Any message or plain object to encode
       * @param [writer] Writer to encode to
       * @returns Writer
       */
      public static encodeDelimited(
        message: google.protobuf.IAny,
        writer?: $protobuf.Writer,
      ): $protobuf.Writer;

      /**
       * Decodes an Any message from the specified reader or buffer.
       * @param reader Reader or buffer to decode from
       * @param [length] Message length if known beforehand
       * @returns Any
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decode(
        reader: $protobuf.Reader | Uint8Array,
        length?: number,
      ): google.protobuf.Any;

      /**
       * Decodes an Any message from the specified reader or buffer, length delimited.
       * @param reader Reader or buffer to decode from
       * @returns Any
       * @throws {Error} If the payload is not a reader or valid buffer
       * @throws {$protobuf.util.ProtocolError} If required fields are missing
       */
      public static decodeDelimited(reader: $protobuf.Reader | Uint8Array): google.protobuf.Any;

      /**
       * Verifies an Any message.
       * @param message Plain object to verify
       * @returns `null` if valid, otherwise the reason why it is not
       */
      public static verify(message: { [k: string]: any }): string | null;

      /**
       * Creates an Any message from a plain object. Also converts values to their respective internal types.
       * @param object Plain object
       * @returns Any
       */
      public static fromObject(object: { [k: string]: any }): google.protobuf.Any;

      /**
       * Creates a plain object from an Any message. Also converts values to other types if specified.
       * @param message Any
       * @param [options] Conversion options
       * @returns Plain object
       */
      public static toObject(
        message: google.protobuf.Any,
        options?: $protobuf.IConversionOptions,
      ): { [k: string]: any };

      /**
       * Converts this Any to JSON.
       * @returns JSON object
       */
      public toJSON(): { [k: string]: any };
    }
  }
}

import 'base.dart';

/// The subset of canonical AG-UI capabilities consumed by Super Manager.
class AgentCapabilities extends AGUIModel {
  final IdentityCapabilities? identity;
  final TransportCapabilities? transport;
  final ToolsCapabilities? tools;
  final StateCapabilities? state;
  final ReasoningCapabilities? reasoning;
  final ExecutionCapabilities? execution;
  final HumanInTheLoopCapabilities? humanInTheLoop;

  const AgentCapabilities({
    this.identity,
    this.transport,
    this.tools,
    this.state,
    this.reasoning,
    this.execution,
    this.humanInTheLoop,
  });

  factory AgentCapabilities.fromJson(Map<String, dynamic> json) =>
      AgentCapabilities(
        identity: _nested(json, 'identity', IdentityCapabilities.fromJson),
        transport: _nested(json, 'transport', TransportCapabilities.fromJson),
        tools: _nested(json, 'tools', ToolsCapabilities.fromJson),
        state: _nested(json, 'state', StateCapabilities.fromJson),
        reasoning: _nested(json, 'reasoning', ReasoningCapabilities.fromJson),
        execution: _nested(json, 'execution', ExecutionCapabilities.fromJson),
        humanInTheLoop: _nested(
          json,
          'humanInTheLoop',
          HumanInTheLoopCapabilities.fromJson,
        ),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (identity != null) 'identity': identity!.toJson(),
        if (transport != null) 'transport': transport!.toJson(),
        if (tools != null) 'tools': tools!.toJson(),
        if (state != null) 'state': state!.toJson(),
        if (reasoning != null) 'reasoning': reasoning!.toJson(),
        if (execution != null) 'execution': execution!.toJson(),
        if (humanInTheLoop != null) 'humanInTheLoop': humanInTheLoop!.toJson(),
      };

  @override
  AgentCapabilities copyWith({
    IdentityCapabilities? identity,
    TransportCapabilities? transport,
    ToolsCapabilities? tools,
    StateCapabilities? state,
    ReasoningCapabilities? reasoning,
    ExecutionCapabilities? execution,
    HumanInTheLoopCapabilities? humanInTheLoop,
  }) =>
      AgentCapabilities(
        identity: identity ?? this.identity,
        transport: transport ?? this.transport,
        tools: tools ?? this.tools,
        state: state ?? this.state,
        reasoning: reasoning ?? this.reasoning,
        execution: execution ?? this.execution,
        humanInTheLoop: humanInTheLoop ?? this.humanInTheLoop,
      );
}

class IdentityCapabilities extends AGUIModel {
  final String? name;
  final String? type;
  final String? version;
  final String? provider;

  const IdentityCapabilities({
    this.name,
    this.type,
    this.version,
    this.provider,
  });

  factory IdentityCapabilities.fromJson(Map<String, dynamic> json) =>
      IdentityCapabilities(
        name: JsonDecoder.optionalField<String>(json, 'name'),
        type: JsonDecoder.optionalField<String>(json, 'type'),
        version: JsonDecoder.optionalField<String>(json, 'version'),
        provider: JsonDecoder.optionalField<String>(json, 'provider'),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (name != null) 'name': name,
        if (type != null) 'type': type,
        if (version != null) 'version': version,
        if (provider != null) 'provider': provider,
      };

  @override
  IdentityCapabilities copyWith({
    String? name,
    String? type,
    String? version,
    String? provider,
  }) =>
      IdentityCapabilities(
        name: name ?? this.name,
        type: type ?? this.type,
        version: version ?? this.version,
        provider: provider ?? this.provider,
      );
}

class TransportCapabilities extends AGUIModel {
  final bool? streaming;
  final bool? resumable;

  const TransportCapabilities({this.streaming, this.resumable});

  factory TransportCapabilities.fromJson(Map<String, dynamic> json) =>
      TransportCapabilities(
        streaming: JsonDecoder.optionalField<bool>(json, 'streaming'),
        resumable: JsonDecoder.optionalField<bool>(json, 'resumable'),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (streaming != null) 'streaming': streaming,
        if (resumable != null) 'resumable': resumable,
      };

  @override
  TransportCapabilities copyWith({bool? streaming, bool? resumable}) =>
      TransportCapabilities(
        streaming: streaming ?? this.streaming,
        resumable: resumable ?? this.resumable,
      );
}

class ToolsCapabilities extends AGUIModel {
  final bool? supported;
  final bool? parallelCalls;
  final bool? clientProvided;

  const ToolsCapabilities({
    this.supported,
    this.parallelCalls,
    this.clientProvided,
  });

  factory ToolsCapabilities.fromJson(Map<String, dynamic> json) =>
      ToolsCapabilities(
        supported: JsonDecoder.optionalField<bool>(json, 'supported'),
        parallelCalls: JsonDecoder.optionalField<bool>(json, 'parallelCalls'),
        clientProvided: JsonDecoder.optionalField<bool>(json, 'clientProvided'),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (supported != null) 'supported': supported,
        if (parallelCalls != null) 'parallelCalls': parallelCalls,
        if (clientProvided != null) 'clientProvided': clientProvided,
      };

  @override
  ToolsCapabilities copyWith({
    bool? supported,
    bool? parallelCalls,
    bool? clientProvided,
  }) =>
      ToolsCapabilities(
        supported: supported ?? this.supported,
        parallelCalls: parallelCalls ?? this.parallelCalls,
        clientProvided: clientProvided ?? this.clientProvided,
      );
}

class StateCapabilities extends AGUIModel {
  final bool? persistentState;

  const StateCapabilities({this.persistentState});

  factory StateCapabilities.fromJson(Map<String, dynamic> json) =>
      StateCapabilities(
        persistentState: JsonDecoder.optionalField<bool>(
          json,
          'persistentState',
        ),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (persistentState != null) 'persistentState': persistentState,
      };

  @override
  StateCapabilities copyWith({bool? persistentState}) => StateCapabilities(
        persistentState: persistentState ?? this.persistentState,
      );
}

class ReasoningCapabilities extends AGUIModel {
  final bool? supported;
  final bool? streaming;
  final bool? encrypted;

  const ReasoningCapabilities({this.supported, this.streaming, this.encrypted});

  factory ReasoningCapabilities.fromJson(Map<String, dynamic> json) =>
      ReasoningCapabilities(
        supported: JsonDecoder.optionalField<bool>(json, 'supported'),
        streaming: JsonDecoder.optionalField<bool>(json, 'streaming'),
        encrypted: JsonDecoder.optionalField<bool>(json, 'encrypted'),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (supported != null) 'supported': supported,
        if (streaming != null) 'streaming': streaming,
        if (encrypted != null) 'encrypted': encrypted,
      };

  @override
  ReasoningCapabilities copyWith({
    bool? supported,
    bool? streaming,
    bool? encrypted,
  }) =>
      ReasoningCapabilities(
        supported: supported ?? this.supported,
        streaming: streaming ?? this.streaming,
        encrypted: encrypted ?? this.encrypted,
      );
}

class ExecutionCapabilities extends AGUIModel {
  final bool? codeExecution;
  final bool? sandboxed;

  const ExecutionCapabilities({this.codeExecution, this.sandboxed});

  factory ExecutionCapabilities.fromJson(Map<String, dynamic> json) =>
      ExecutionCapabilities(
        codeExecution: JsonDecoder.optionalField<bool>(json, 'codeExecution'),
        sandboxed: JsonDecoder.optionalField<bool>(json, 'sandboxed'),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (codeExecution != null) 'codeExecution': codeExecution,
        if (sandboxed != null) 'sandboxed': sandboxed,
      };

  @override
  ExecutionCapabilities copyWith({bool? codeExecution, bool? sandboxed}) =>
      ExecutionCapabilities(
        codeExecution: codeExecution ?? this.codeExecution,
        sandboxed: sandboxed ?? this.sandboxed,
      );
}

class HumanInTheLoopCapabilities extends AGUIModel {
  final bool? supported;
  final bool? approvals;
  final bool? interventions;
  final bool? interrupts;
  final bool? approveWithEdits;

  const HumanInTheLoopCapabilities({
    this.supported,
    this.approvals,
    this.interventions,
    this.interrupts,
    this.approveWithEdits,
  });

  factory HumanInTheLoopCapabilities.fromJson(Map<String, dynamic> json) =>
      HumanInTheLoopCapabilities(
        supported: JsonDecoder.optionalField<bool>(json, 'supported'),
        approvals: JsonDecoder.optionalField<bool>(json, 'approvals'),
        interventions: JsonDecoder.optionalField<bool>(json, 'interventions'),
        interrupts: JsonDecoder.optionalField<bool>(json, 'interrupts'),
        approveWithEdits: JsonDecoder.optionalField<bool>(
          json,
          'approveWithEdits',
        ),
      );

  @override
  Map<String, dynamic> toJson() => {
        if (supported != null) 'supported': supported,
        if (approvals != null) 'approvals': approvals,
        if (interventions != null) 'interventions': interventions,
        if (interrupts != null) 'interrupts': interrupts,
        if (approveWithEdits != null) 'approveWithEdits': approveWithEdits,
      };

  @override
  HumanInTheLoopCapabilities copyWith({
    bool? supported,
    bool? approvals,
    bool? interventions,
    bool? interrupts,
    bool? approveWithEdits,
  }) =>
      HumanInTheLoopCapabilities(
        supported: supported ?? this.supported,
        approvals: approvals ?? this.approvals,
        interventions: interventions ?? this.interventions,
        interrupts: interrupts ?? this.interrupts,
        approveWithEdits: approveWithEdits ?? this.approveWithEdits,
      );
}

T? _nested<T>(
  Map<String, dynamic> json,
  String field,
  T Function(Map<String, dynamic>) decode,
) {
  final value = JsonDecoder.optionalField<Map<String, dynamic>>(json, field);
  return value == null ? null : decode(value);
}

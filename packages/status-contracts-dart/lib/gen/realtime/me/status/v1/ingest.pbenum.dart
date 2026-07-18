// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/ingest.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ScrapeJob is the Prometheus job a scrape target belongs to.
class ScrapeJob extends $pb.ProtobufEnum {
  /// Job is not known.
  static const ScrapeJob SCRAPE_JOB_UNSPECIFIED =
      ScrapeJob._(0, _omitEnumNames ? '' : 'SCRAPE_JOB_UNSPECIFIED');

  /// Linux node-exporter on a physical host.
  static const ScrapeJob SCRAPE_JOB_NODE_EXPORTER =
      ScrapeJob._(1, _omitEnumNames ? '' : 'SCRAPE_JOB_NODE_EXPORTER');

  /// Linux node-exporter on a virtual machine.
  static const ScrapeJob SCRAPE_JOB_VM_NODE_EXPORTER =
      ScrapeJob._(2, _omitEnumNames ? '' : 'SCRAPE_JOB_VM_NODE_EXPORTER');

  /// Device exporter for media and accessory signals.
  static const ScrapeJob SCRAPE_JOB_DEVICE_EXPORTER =
      ScrapeJob._(3, _omitEnumNames ? '' : 'SCRAPE_JOB_DEVICE_EXPORTER');

  /// Coding-agent exporter.
  static const ScrapeJob SCRAPE_JOB_AGENT_EXPORTER =
      ScrapeJob._(4, _omitEnumNames ? '' : 'SCRAPE_JOB_AGENT_EXPORTER');

  static const $core.List<ScrapeJob> values = <ScrapeJob>[
    SCRAPE_JOB_UNSPECIFIED,
    SCRAPE_JOB_NODE_EXPORTER,
    SCRAPE_JOB_VM_NODE_EXPORTER,
    SCRAPE_JOB_DEVICE_EXPORTER,
    SCRAPE_JOB_AGENT_EXPORTER,
  ];

  static final $core.List<ScrapeJob?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static ScrapeJob? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ScrapeJob._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');

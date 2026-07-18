/// AG-UI protocol models and streaming primitives used by Realtime Me Manager.
library ag_ui;

export 'src/events/events.dart';
export 'src/sse/sse_message.dart';
export 'src/sse/sse_parser.dart';
export 'src/types/types.dart';

/// Version of the vendored protocol package.
const String agUiVersion = '0.3.0+realtimeme.1';

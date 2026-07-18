import { type AGUIEvent, EventType } from "@ag-ui/core";

export function assistantMessageStart(messageId: string): AGUIEvent {
  return { type: EventType.TEXT_MESSAGE_START, messageId, role: "assistant" };
}

export function assistantMessageDelta(messageId: string, delta: string): AGUIEvent {
  return { type: EventType.TEXT_MESSAGE_CONTENT, messageId, delta };
}

export function assistantMessageEnd(messageId: string): AGUIEvent {
  return { type: EventType.TEXT_MESSAGE_END, messageId };
}

export function toolCallStart(
  toolCallId: string,
  name: string,
  input: unknown,
): readonly AGUIEvent[] {
  return [
    { type: EventType.TOOL_CALL_START, toolCallId, toolCallName: name },
    { type: EventType.TOOL_CALL_ARGS, toolCallId, delta: JSON.stringify(input) },
    { type: EventType.TOOL_CALL_END, toolCallId },
  ];
}

export function toolCallResult(toolCallId: string, content: unknown): AGUIEvent {
  return {
    type: EventType.TOOL_CALL_RESULT,
    toolCallId,
    messageId: `result-${toolCallId}`,
    content: typeof content === "string" ? content : JSON.stringify(content),
    role: "tool",
  };
}

export function activityDelta(
  messageId: string,
  activityType: string,
  content: Record<string, unknown>,
): AGUIEvent {
  return {
    type: EventType.ACTIVITY_SNAPSHOT,
    messageId,
    activityType,
    content,
    replace: false,
  };
}

export function reasoningStart(messageId: string): AGUIEvent {
  return { type: EventType.REASONING_MESSAGE_START, messageId, role: "reasoning" };
}

export function reasoningDelta(messageId: string, delta: string): AGUIEvent {
  return { type: EventType.REASONING_MESSAGE_CONTENT, messageId, delta };
}

export function reasoningEnd(messageId: string): AGUIEvent {
  return { type: EventType.REASONING_MESSAGE_END, messageId };
}

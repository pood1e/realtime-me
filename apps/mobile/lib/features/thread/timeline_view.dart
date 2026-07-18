import 'package:flutter/material.dart';
import 'package:flutter_markdown_plus/flutter_markdown_plus.dart';

import '../../ui/common.dart';
import 'timeline_projection.dart';

class TimelineView extends StatelessWidget {
  final List<TimelineItem> items;
  final ScrollController controller;

  const TimelineView({
    required this.items,
    required this.controller,
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const EmptyState(
        icon: Icons.chat_bubble_outline_rounded,
        title: '开始一项任务',
        message: '输入目标后，Codex 或 Claude Code 的回复、工具调用和提问会实时显示在这里。',
      );
    }
    return ListView.builder(
      controller: controller,
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
      itemCount: items.length,
      itemBuilder: (context, index) => Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: _TimelineItemView(item: items[index]),
      ),
    );
  }
}

class _TimelineItemView extends StatelessWidget {
  final TimelineItem item;

  const _TimelineItemView({required this.item});

  @override
  Widget build(BuildContext context) {
    return switch (item) {
      final TimelineMessageItem message => _MessageView(message: message),
      final TimelineToolItem tool => _ToolView(tool: tool),
      final TimelineReasoningItem reasoning => _ReasoningView(
        reasoning: reasoning,
      ),
      final TimelineActivityItem activity => _ActivityView(activity: activity),
      final TimelineNoticeItem notice => _NoticeView(notice: notice),
    };
  }
}

class _MessageView extends StatelessWidget {
  final TimelineMessageItem message;

  const _MessageView({required this.message});

  @override
  Widget build(BuildContext context) {
    final user = message.role == TimelineMessageRole.user;
    final scheme = Theme.of(context).colorScheme;
    return Align(
      alignment: user ? Alignment.centerRight : Alignment.centerLeft,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 760),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: user ? scheme.primaryContainer : scheme.surfaceContainerLow,
            borderRadius: BorderRadius.circular(16),
            border: user ? null : Border.all(color: scheme.outlineVariant),
          ),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            child: message.content.isEmpty
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : MarkdownBody(
                    data: message.content,
                    selectable: true,
                    softLineBreak: true,
                  ),
          ),
        ),
      ),
    );
  }
}

class _ToolView extends StatelessWidget {
  final TimelineToolItem tool;

  const _ToolView({required this.tool});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: ExpansionTile(
        leading: Icon(
          tool.complete ? Icons.build_circle_outlined : Icons.sync_rounded,
        ),
        title: Text(tool.name),
        subtitle: Text(tool.complete ? '工具调用完成' : '正在执行'),
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        children: [
          if (tool.arguments.isNotEmpty)
            _CodeSection(title: '参数', content: tool.arguments),
          if (tool.result.isNotEmpty) ...[
            if (tool.arguments.isNotEmpty) const SizedBox(height: 12),
            _CodeSection(title: '结果', content: tool.result),
          ],
        ],
      ),
    );
  }
}

class _ReasoningView extends StatelessWidget {
  final TimelineReasoningItem reasoning;

  const _ReasoningView({required this.reasoning});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: ExpansionTile(
        leading: const Icon(Icons.psychology_alt_outlined),
        title: const Text('推理摘要'),
        subtitle: Text(reasoning.complete ? '已完成' : '生成中'),
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        children: [
          Align(
            alignment: Alignment.centerLeft,
            child: MarkdownBody(
              data: reasoning.content.isEmpty ? '等待摘要…' : reasoning.content,
              selectable: true,
              softLineBreak: true,
            ),
          ),
        ],
      ),
    );
  }
}

class _ActivityView extends StatelessWidget {
  final TimelineActivityItem activity;

  const _ActivityView({required this.activity});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: ExpansionTile(
        leading: const Icon(Icons.track_changes_rounded),
        title: Text(_activityTitle(activity.activityType)),
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        children: [_CodeSection(content: activity.displayContent)],
      ),
    );
  }
}

class _NoticeView extends StatelessWidget {
  final TimelineNoticeItem notice;

  const _NoticeView({required this.notice});

  @override
  Widget build(BuildContext context) {
    final error = notice.kind == TimelineNoticeKind.error;
    final scheme = Theme.of(context).colorScheme;
    return Card(
      color: error ? scheme.errorContainer : scheme.secondaryContainer,
      child: ListTile(
        leading: Icon(
          error
              ? Icons.error_outline_rounded
              : Icons.subdirectory_arrow_right_rounded,
          color: error ? scheme.onErrorContainer : scheme.onSecondaryContainer,
        ),
        title: Text(notice.title),
        subtitle: notice.message.isEmpty
            ? null
            : SelectableText(notice.message),
      ),
    );
  }
}

class _CodeSection extends StatelessWidget {
  final String? title;
  final String content;

  const _CodeSection({this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Align(
      alignment: Alignment.centerLeft,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (title != null) ...[
            Text(title!, style: theme.textTheme.labelLarge),
            const SizedBox(height: 6),
          ],
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(10),
            ),
            child: SelectableText(
              content,
              style: theme.textTheme.bodySmall?.copyWith(
                fontFamily: 'monospace',
                height: 1.45,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

String _activityTitle(String type) => switch (type) {
  'plan' => '执行计划',
  'tool_progress' => '工具进度',
  'output_truncated' => '输出已截断',
  _ => type,
};

import 'package:ag_ui/ag_ui.dart' hide State;
import 'package:flutter/material.dart';

import '../../ui/common.dart';
import 'timeline_projection.dart';

typedef _AnswerKey = ({String interruptId, String questionId});

class InterruptPanel extends StatefulWidget {
  final PendingInterruptBatch batch;
  final Future<void> Function(List<ResumeEntry> entries) onSubmit;

  const InterruptPanel({
    required this.batch,
    required this.onSubmit,
    super.key,
  });

  @override
  State<InterruptPanel> createState() => _InterruptPanelState();
}

class _InterruptPanelState extends State<InterruptPanel> {
  final _controllers = <_AnswerKey, TextEditingController>{};
  final _otherControllers = <_AnswerKey, TextEditingController>{};
  final _selections = <_AnswerKey, Set<String>>{};
  final _otherSelected = <_AnswerKey, bool>{};
  final _obscured = <_AnswerKey, bool>{};
  late final List<_InterruptQuestions> _groups;
  String? _validationError;
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _groups = _parseGroups(widget.batch.interrupts);
    for (final group in _groups) {
      for (final question in group.questions) {
        final key = (interruptId: group.interrupt.id, questionId: question.id);
        _selections[key] = <String>{};
        _otherSelected[key] = false;
        _obscured[key] = question.secret;
        if (question.options.isEmpty) {
          _controllers[key] = TextEditingController();
        }
        if (question.allowOther) {
          _otherControllers[key] = TextEditingController();
        }
      }
    }
  }

  @override
  void dispose() {
    for (final controller in [
      ..._controllers.values,
      ..._otherControllers.values,
    ]) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Material(
      color: scheme.tertiaryContainer,
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 420),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.question_answer_outlined,
                      color: scheme.onTertiaryContainer,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        '需要你的输入',
                        style: Theme.of(context).textTheme.titleMedium
                            ?.copyWith(
                              color: scheme.onTertiaryContainer,
                              fontWeight: FontWeight.w700,
                            ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: _groups.isEmpty
                      ? const Center(child: Text('服务端返回的提问格式无法识别，请停止本次任务。'))
                      : ListView.separated(
                          itemCount: _groups.length,
                          separatorBuilder: (context, index) =>
                              const Divider(height: 28),
                          itemBuilder: (context, index) =>
                              _buildGroup(_groups[index]),
                        ),
                ),
                if (_validationError != null) ...[
                  const SizedBox(height: 8),
                  Text(
                    _validationError!,
                    style: TextStyle(
                      color: scheme.error,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
                const SizedBox(height: 12),
                Align(
                  alignment: Alignment.centerRight,
                  child: FilledButton.icon(
                    onPressed: _groups.isEmpty || _submitting ? null : _submit,
                    icon: _submitting
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.send_rounded),
                    label: const Text('提交回答'),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildGroup(_InterruptQuestions group) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (var index = 0; index < group.questions.length; index++) ...[
          if (index > 0) const SizedBox(height: 20),
          _buildQuestion(group.interrupt.id, group.questions[index]),
        ],
      ],
    );
  }

  Widget _buildQuestion(String interruptId, _Question question) {
    final key = (interruptId: interruptId, questionId: question.id);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (question.header.isNotEmpty) ...[
          Text(question.header, style: Theme.of(context).textTheme.labelLarge),
          const SizedBox(height: 4),
        ],
        Text(question.question, style: Theme.of(context).textTheme.bodyLarge),
        const SizedBox(height: 10),
        if (question.options.isEmpty)
          _answerField(
            key: key,
            controller: _controllers[key]!,
            label: question.secret ? '敏感回答' : '回答',
            secret: question.secret,
          )
        else
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final option in question.options)
                question.multiple
                    ? FilterChip(
                        label: Text(option),
                        selected: _selections[key]!.contains(option),
                        onSelected: _submitting
                            ? null
                            : (selected) => _toggleOption(
                                key,
                                option,
                                selected,
                                multiple: true,
                              ),
                      )
                    : ChoiceChip(
                        label: Text(option),
                        selected: _selections[key]!.contains(option),
                        onSelected: _submitting
                            ? null
                            : (selected) => _toggleOption(
                                key,
                                option,
                                selected,
                                multiple: false,
                              ),
                      ),
              if (question.allowOther)
                question.multiple
                    ? FilterChip(
                        label: const Text('其他'),
                        selected: _otherSelected[key]!,
                        onSelected: _submitting
                            ? null
                            : (selected) => setState(
                                () => _otherSelected[key] = selected,
                              ),
                      )
                    : ChoiceChip(
                        label: const Text('其他'),
                        selected: _otherSelected[key]!,
                        onSelected: _submitting
                            ? null
                            : (selected) {
                                setState(() {
                                  _otherSelected[key] = selected;
                                  if (selected) {
                                    _selections[key]!.clear();
                                  }
                                });
                              },
                      ),
            ],
          ),
        if (question.allowOther && _otherSelected[key]!) ...[
          const SizedBox(height: 10),
          _answerField(
            key: key,
            controller: _otherControllers[key]!,
            label: '其他回答',
            secret: question.secret,
          ),
        ],
        if (question.multiple && question.options.isNotEmpty) ...[
          const SizedBox(height: 6),
          Text('可多选', style: Theme.of(context).textTheme.bodySmall),
        ],
      ],
    );
  }

  Widget _answerField({
    required _AnswerKey key,
    required TextEditingController controller,
    required String label,
    required bool secret,
  }) {
    return TextField(
      controller: controller,
      enabled: !_submitting,
      obscureText: secret && _obscured[key]!,
      minLines: 1,
      maxLines: secret && _obscured[key]! ? 1 : 4,
      textInputAction: TextInputAction.done,
      decoration: InputDecoration(
        labelText: label,
        suffixIcon: secret
            ? IconButton(
                tooltip: _obscured[key]! ? '显示回答' : '隐藏回答',
                onPressed: _submitting
                    ? null
                    : () => setState(() => _obscured[key] = !_obscured[key]!),
                icon: Icon(
                  _obscured[key]!
                      ? Icons.visibility_outlined
                      : Icons.visibility_off_outlined,
                ),
              )
            : null,
      ),
    );
  }

  void _toggleOption(
    _AnswerKey key,
    String option,
    bool selected, {
    required bool multiple,
  }) {
    setState(() {
      final values = _selections[key]!;
      if (!multiple) {
        values.clear();
        _otherSelected[key] = false;
      }
      if (selected) {
        values.add(option);
      } else {
        values.remove(option);
      }
    });
  }

  Future<void> _submit() async {
    final entries = <ResumeEntry>[];
    for (final group in _groups) {
      final answers = <String, List<String>>{};
      for (final question in group.questions) {
        final key = (interruptId: group.interrupt.id, questionId: question.id);
        final values = <String>[];
        if (question.options.isEmpty) {
          final value = _controllers[key]!.text.trim();
          if (value.isNotEmpty) {
            values.add(value);
          }
        } else {
          values.addAll(question.options.where(_selections[key]!.contains));
        }
        if (question.allowOther && _otherSelected[key]!) {
          final other = _otherControllers[key]!.text.trim();
          if (other.isNotEmpty) {
            values.add(other);
          }
        }
        if (values.isEmpty || (!question.multiple && values.length != 1)) {
          setState(() => _validationError = '请完整回答“${question.question}”。');
          return;
        }
        answers[question.id] = values;
      }
      entries.add(
        ResumeEntry(
          interruptId: group.interrupt.id,
          status: ResumeStatus.resolved,
          payload: {'answers': answers},
        ),
      );
    }
    setState(() {
      _validationError = null;
      _submitting = true;
    });
    try {
      await widget.onSubmit(entries);
      for (final group in _groups) {
        for (final question in group.questions.where(
          (question) => question.secret,
        )) {
          final key = (
            interruptId: group.interrupt.id,
            questionId: question.id,
          );
          _controllers[key]?.clear();
          _otherControllers[key]?.clear();
        }
      }
    } on Object catch (error) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _validationError = readableError(error);
        });
      }
    }
  }
}

final class _InterruptQuestions {
  final Interrupt interrupt;
  final List<_Question> questions;

  const _InterruptQuestions({required this.interrupt, required this.questions});
}

final class _Question {
  final String id;
  final String header;
  final String question;
  final List<String> options;
  final bool multiple;
  final bool secret;
  final bool allowOther;

  const _Question({
    required this.id,
    required this.header,
    required this.question,
    required this.options,
    required this.multiple,
    required this.secret,
    required this.allowOther,
  });
}

List<_InterruptQuestions> _parseGroups(List<Interrupt> interrupts) {
  final groups = <_InterruptQuestions>[];
  for (final interrupt in interrupts) {
    final rawQuestions = interrupt.metadata?['questions'];
    if (rawQuestions is! List<dynamic>) {
      return const [];
    }
    final questions = <_Question>[];
    for (final raw in rawQuestions) {
      if (raw is! Map<String, dynamic>) {
        return const [];
      }
      final id = raw['id'];
      final question = raw['question'];
      final options = raw['options'];
      if (id is! String ||
          id.isEmpty ||
          question is! String ||
          question.isEmpty) {
        return const [];
      }
      if (options is! List<dynamic> ||
          options.any((option) => option is! String)) {
        return const [];
      }
      questions.add(
        _Question(
          id: id,
          header: raw['header'] is String ? raw['header']! as String : '',
          question: question,
          options: options.cast<String>(),
          multiple: raw['multiple'] == true,
          secret: raw['secret'] == true,
          allowOther: raw['allowOther'] == true,
        ),
      );
    }
    if (questions.isEmpty) {
      return const [];
    }
    groups.add(_InterruptQuestions(interrupt: interrupt, questions: questions));
  }
  return groups;
}

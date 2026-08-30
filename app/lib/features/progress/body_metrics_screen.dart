import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';

final bodyMetricsProvider = FutureProvider.autoDispose<List<BodyMetric>>((ref) => ref.watch(apiRepositoryProvider).bodyMetrics());

class BodyMetricsScreen extends ConsumerWidget {
  const BodyMetricsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final metricsAsync = ref.watch(bodyMetricsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Peso y medidas')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showLogSheet(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Registrar'),
      ),
      body: metricsAsync.when(
        data: (metrics) {
          if (metrics.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('Todavía no registraste tu peso. Usá el botón de abajo.', textAlign: TextAlign.center),
              ),
            );
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _WeightChart(metrics: metrics),
              const SizedBox(height: 16),
              ...metrics.map((m) => _MetricRow(metric: m)),
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => const Center(child: Text('No se pudieron cargar tus medidas')),
      ),
    );
  }

  void _showLogSheet(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
        child: const _LogMetricForm(),
      ),
    );
  }
}

class _LogMetricForm extends ConsumerStatefulWidget {
  const _LogMetricForm();

  @override
  ConsumerState<_LogMetricForm> createState() => _LogMetricFormState();
}

class _LogMetricFormState extends ConsumerState<_LogMetricForm> {
  final _weightCtrl = TextEditingController();
  final _fatCtrl = TextEditingController();
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _weightCtrl.dispose();
    _fatCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final weight = double.tryParse(_weightCtrl.text.replaceAll(',', '.'));
    final fat = double.tryParse(_fatCtrl.text.replaceAll(',', '.'));
    if (weight == null && fat == null) {
      setState(() => _error = 'Ingresá al menos el peso.');
      return;
    }
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      await ref.read(apiRepositoryProvider).logBodyMetric(
            measuredOn: DateFormat('yyyy-MM-dd').format(DateTime.now()),
            weightKg: weight,
            bodyFatPct: fat,
          );
      ref.invalidate(bodyMetricsProvider);
      if (mounted) Navigator.pop(context);
    } catch (_) {
      if (mounted) setState(() => _error = 'No se pudo guardar. Probá de nuevo.');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('Registrar medida de hoy', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 16),
          TextField(
            controller: _weightCtrl,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(labelText: 'Peso (kg)'),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _fatCtrl,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(labelText: '% de grasa (opcional)'),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: TextStyle(color: colors.danger)),
          ],
          const SizedBox(height: 20),
          FilledButton(
            onPressed: _saving ? null : _save,
            child: _saving ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Guardar'),
          ),
        ],
      ),
    );
  }
}

class _WeightChart extends StatelessWidget {
  const _WeightChart({required this.metrics});
  final List<BodyMetric> metrics;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    // El endpoint devuelve lo mas reciente primero; la grafica va al reves.
    final withWeight = metrics.where((m) => m.weightKg != null).toList().reversed.toList();
    if (withWeight.length < 2) return const SizedBox.shrink();

    final spots = [for (var i = 0; i < withWeight.length; i++) FlSpot(i.toDouble(), withWeight[i].weightKg!)];
    final latest = withWeight.last.weightKg!;
    final first = withWeight.first.weightKg!;
    final delta = latest - first;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('${latest.toStringAsFixed(1)} kg', style: Theme.of(context).textTheme.headlineMedium),
                Text(
                  '${delta >= 0 ? '+' : ''}${delta.toStringAsFixed(1)} kg',
                  style: TextStyle(color: delta >= 0 ? colors.warning : colors.success, fontWeight: FontWeight.w700),
                ),
              ],
            ),
            SizedBox(
              height: 140,
              child: Padding(
                padding: const EdgeInsets.only(top: 12),
                child: LineChart(LineChartData(
                  gridData: const FlGridData(show: false),
                  titlesData: const FlTitlesData(show: false),
                  borderData: FlBorderData(show: false),
                  lineBarsData: [
                    LineChartBarData(spots: spots, isCurved: true, color: colors.brand, barWidth: 3, dotData: const FlDotData(show: false)),
                  ],
                )),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MetricRow extends StatelessWidget {
  const _MetricRow({required this.metric});
  final BodyMetric metric;

  @override
  Widget build(BuildContext context) {
    final parts = [
      if (metric.weightKg != null) '${metric.weightKg!.toStringAsFixed(1)} kg',
      if (metric.bodyFatPct != null) '${metric.bodyFatPct!.toStringAsFixed(1)}% grasa',
    ];
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        title: Text(parts.join(' · ')),
        subtitle: Text(metric.measuredOn),
      ),
    );
  }
}

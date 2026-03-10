package report

const sweepHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <style>
    body {
      font-family: Arial, sans-serif;
      margin: 32px;
      color: #222;
    }
    h1, h2 {
      margin-bottom: 8px;
    }
    .meta {
      margin-bottom: 24px;
      padding: 12px 16px;
      background: #f5f5f5;
      border-radius: 8px;
    }
    .chart {
      margin-bottom: 36px;
    }
    canvas {
      max-width: 900px;
      max-height: 420px;
    }
  </style>
</head>
<body>
  <h1>llmbench Sweep Report</h1>

  <div class="meta">
    <div><strong>Model:</strong> {{.Model}}</div>
    <div><strong>URL:</strong> {{.URL}}</div>
    <div><strong>Prompt tokens:</strong> {{.PromptTokens}}</div>
    <div><strong>Completion tokens:</strong> {{.CompletionTokens}}</div>
    <div><strong>Requests per step:</strong> {{.Requests}}</div>
  </div>

  <div class="chart">
    <h2>Output Tokens / Second</h2>
    <canvas id="tokChart"></canvas>
  </div>

  <div class="chart">
    <h2>Average Latency (ms)</h2>
    <canvas id="avgLatencyChart"></canvas>
  </div>

  <div class="chart">
    <h2>Latency P95 (ms)</h2>
    <canvas id="latencyP95Chart"></canvas>
  </div>

  <div class="chart">
    <h2>TTFT P50 (ms)</h2>
    <canvas id="ttftP50Chart"></canvas>
  </div>

  <script>
    const labels = {{.ConcurrencyJSON}};
    const tokPerSec = {{.TokensPerSecJSON}};
    const avgLatency = {{.AvgLatencyJSON}};
    const latencyP95 = {{.LatencyP95JSON}};
    const ttftP50 = {{.TTFTP50JSON}};

    function makeLineChart(id, label, data) {
      new Chart(document.getElementById(id), {
        type: 'line',
        data: {
          labels,
          datasets: [{
            label,
            data,
            tension: 0.2
          }]
        },
        options: {
          responsive: true,
          scales: {
            x: { title: { display: true, text: 'Concurrency' } },
            y: { beginAtZero: true }
          }
        }
      });
    }

    makeLineChart('tokChart', 'Output tokens/sec', tokPerSec);
    makeLineChart('avgLatencyChart', 'Average latency (ms)', avgLatency);
    makeLineChart('latencyP95Chart', 'Latency p95 (ms)', latencyP95);
    makeLineChart('ttftP50Chart', 'TTFT p50 (ms)', ttftP50);
  </script>
</body>
</html>`

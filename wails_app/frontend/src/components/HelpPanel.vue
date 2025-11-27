<script setup>
import {ref} from 'vue'

const activeSection = ref('intro')
</script>

<template>
  <div class="help-panel">
    <div class="card help-card">
      <div class="card-header">
        <h2>{{ $t('help.title') }}</h2>
      </div>
      
      <div class="help-content">
        <section class="help-section">
          <h3>{{ $t('help.intro') }}</h3>
        </section>

        <section class="help-section">
          <h3>{{ $t('help.quick_start.title') }}</h3>
          <ol>
            <li>{{ $t('help.quick_start.step1') }}</li>
            <li>{{ $t('help.quick_start.step2') }}</li>
            <li>{{ $t('help.quick_start.step3') }}</li>
          </ol>
        </section>

        <section class="help-section">
          <h3>{{ $t('help.config.title') }}</h3>
          <p>{{ $t('help.config.desc') }}</p>
          <div class="info-box">
            {{ $t('help.config.proxy') }}
          </div>
        </section>

        <section class="help-section">
          <h3>{{ $t('help.faq.title') }}</h3>
          <div class="faq-item">
            <div class="question">Q: {{ $t('help.faq.q1') }}</div>
            <div class="answer">A: {{ $t('help.faq.a1') }}</div>
          </div>
          <div class="faq-item">
            <div class="question">Q: {{ $t('help.faq.q2') }}</div>
            <div class="answer">A: {{ $t('help.faq.a2') }}</div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.help-panel {
  max-width: 1000px;
  margin: 0 auto;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.card-header {
  padding: 24px;
  border-bottom: 1px solid var(--border-color);
}

.card-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--text-primary);
}

.help-content {
  padding: 24px;
}

.help-section {
  margin-bottom: 32px;
}

.help-section:last-child {
  margin-bottom: 0;
}

.help-section h3 {
  font-size: 16px;
  color: var(--text-primary);
  margin: 0 0 16px 0;
  font-weight: 600;
}

ol {
  margin: 0;
  padding-left: 20px;
  color: var(--text-secondary);
  line-height: 1.6;
}

li {
  margin-bottom: 8px;
}

p {
  margin: 0 0 12px 0;
  color: var(--text-secondary);
  line-height: 1.6;
}

.info-box {
  background: #e3f2fd;
  color: #1565c0;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 14px;
  border-left: 4px solid #1565c0;
}

.faq-item {
  margin-bottom: 16px;
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
}

.question {
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.answer {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.6;
}
</style>

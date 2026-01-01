/**
 * Memory Management System
 * Monitors JavaScript heap memory usage and triggers shutdown when limit is exceeded
 * Prevents browser crashes and freezes by graceful termination
 */

import {
  MAX_MEMORY_MB,
  MEMORY_WARNING_THRESHOLD,
  MEMORY_CHECK_INTERVAL,
} from "../utils/constants.js";

export class MemoryManager {
  constructor(options = {}) {
    this.maxMemoryMB = options.maxMemoryMB || MAX_MEMORY_MB || 512;
    this.warningThreshold =
      options.warningThreshold || MEMORY_WARNING_THRESHOLD || 0.8;
    this.checkInterval = options.checkInterval || MEMORY_CHECK_INTERVAL || 2000;

    // Callbacks
    this.onWarning = options.onWarning || null;
    this.onCritical = options.onCritical || null;
    this.onUpdate = options.onUpdate || null;

    this.intervalId = null;
    this.isRunning = false;
    this.warningTriggered = false;

    // Check if memory API is available
    this.hasMemoryAPI =
      typeof performance !== "undefined" && performance.memory !== undefined;
  }

  /**
   * Start memory monitoring
   */
  start() {
    if (this.isRunning) return;

    this.isRunning = true;
    this.warningTriggered = false;

    // Initial check
    this._check();

    // Periodic checks
    this.intervalId = setInterval(() => {
      this._check();
    }, this.checkInterval);

    console.log(
      `[MemoryManager] Started monitoring (limit: ${this.maxMemoryMB}MB)`
    );
  }

  /**
   * Stop memory monitoring
   */
  stop() {
    if (!this.isRunning) return;

    this.isRunning = false;

    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }

    console.log("[MemoryManager] Stopped monitoring");
  }

  /**
   * Get current memory usage statistics
   * @returns {Object} Memory usage info
   */
  getUsage() {
    if (!this.hasMemoryAPI) {
      // Fallback for browsers without memory API
      return {
        used: 0,
        limit: this.maxMemoryMB,
        percentage: 0,
        available: true,
        supported: false,
      };
    }

    const memory = performance.memory;
    const usedMB = Math.round(memory.usedJSHeapSize / (1024 * 1024));
    const totalMB = Math.round(memory.totalJSHeapSize / (1024 * 1024));
    const limitMB = Math.round(memory.jsHeapSizeLimit / (1024 * 1024));

    return {
      used: usedMB,
      total: totalMB,
      heapLimit: limitMB,
      limit: this.maxMemoryMB,
      percentage: usedMB / this.maxMemoryMB,
      available: true,
      supported: true,
    };
  }

  /**
   * Force garbage collection hint (not guaranteed)
   */
  requestGC() {
    // Attempt to reduce memory pressure
    if (typeof window !== "undefined" && window.gc) {
      window.gc();
    }
  }

  /**
   * Internal check method
   */
  _check() {
    const usage = this.getUsage();

    // Notify update callback
    if (this.onUpdate) {
      this.onUpdate(usage);
    }

    if (!usage.supported) {
      return; // Can't monitor without API
    }

    const percentage = usage.percentage;

    // Critical - exceeds limit
    if (percentage >= 1.0) {
      console.error(
        `[MemoryManager] CRITICAL: Memory limit exceeded (${usage.used}MB / ${usage.limit}MB)`
      );

      if (this.onCritical) {
        this.stop(); // Stop monitoring before shutdown
        this.onCritical(usage);
      }
      return;
    }

    // Warning threshold
    if (percentage >= this.warningThreshold && !this.warningTriggered) {
      this.warningTriggered = true;
      console.warn(
        `[MemoryManager] WARNING: Memory usage high (${usage.used}MB / ${
          usage.limit
        }MB - ${Math.round(percentage * 100)}%)`
      );

      if (this.onWarning) {
        this.onWarning(usage);
      }

      // Try to free some memory
      this.requestGC();
    }

    // Reset warning if usage drops
    if (percentage < this.warningThreshold * 0.9) {
      this.warningTriggered = false;
    }
  }

  /**
   * Get formatted string for display
   */
  getDisplayString() {
    const usage = this.getUsage();

    if (!usage.supported) {
      return "N/A";
    }

    return `${usage.used}MB / ${usage.limit}MB`;
  }
}

export default MemoryManager;

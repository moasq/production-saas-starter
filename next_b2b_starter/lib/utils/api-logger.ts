/**
 * API Request/Response Logger
 *
 * Provides comprehensive logging for all API requests and responses
 * with color-coded console output for easy debugging.
 */

export interface LoggedRequest {
  method: string;
  url: string;
  headers: Record<string, string>;
  body?: any;
  timestamp: number;
}

export interface LoggedResponse {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body?: any;
  duration: number;
  timestamp: number;
}

export interface LoggedError {
  message: string;
  status?: number;
  details?: any;
  duration: number;
  timestamp: number;
}

class ApiLogger {
  private enabled: boolean;
  private pendingRequests: Map<string, number>;

  constructor() {
    // Enable in development or when explicitly enabled
    this.enabled =
      typeof window !== "undefined" &&
      (process.env.NODE_ENV === "development" ||
        localStorage.getItem("API_LOGGER_ENABLED") === "true");

    this.pendingRequests = new Map();
  }

  /**
   * Enable/disable API logging
   */
  setEnabled(enabled: boolean): void {
    this.enabled = enabled;
    if (typeof window !== "undefined") {
      if (enabled) {
        localStorage.setItem("API_LOGGER_ENABLED", "true");
      } else {
        localStorage.removeItem("API_LOGGER_ENABLED");
      }
    }
  }

  /**
   * Log an outgoing API request
   */
  logRequest(requestId: string, request: LoggedRequest): void {
    if (!this.enabled) return;

    this.pendingRequests.set(requestId, request.timestamp);

    const style = "color: #2196F3; font-weight: bold;";
    const resetStyle = "color: inherit; font-weight: normal;";

    console.groupCollapsed(
      `%c→ ${request.method} %c${request.url}`,
      style,
      resetStyle
    );

    // Log request details
    console.log("%cRequest Details:", "font-weight: bold;");
    console.table({
      Method: request.method,
      URL: request.url,
      Timestamp: new Date(request.timestamp).toISOString(),
    });

    // Log headers
    console.log("%cHeaders:", "font-weight: bold; color: #9C27B0;");
    console.table(this.sanitizeHeaders(request.headers));

    // Log body if present
    if (request.body) {
      console.log("%cBody:", "font-weight: bold; color: #FF9800;");
      try {
        const parsed = typeof request.body === "string"
          ? JSON.parse(request.body)
          : request.body;
        console.log(parsed);
      } catch {
        console.log(request.body);
      }
    }

    console.groupEnd();
  }

  /**
   * Log a successful API response
   */
  logResponse(requestId: string, response: LoggedResponse): void {
    if (!this.enabled) return;

    const startTime = this.pendingRequests.get(requestId);
    const duration = startTime ? response.timestamp - startTime : response.duration;
    this.pendingRequests.delete(requestId);

    const isSuccess = response.status >= 200 && response.status < 300;
    const style = isSuccess
      ? "color: #4CAF50; font-weight: bold;"
      : "color: #FF5722; font-weight: bold;";
    const resetStyle = "color: inherit; font-weight: normal;";

    console.groupCollapsed(
      `%c← ${response.status} %c${response.statusText} %c(${duration}ms)`,
      style,
      resetStyle,
      "color: #757575; font-size: 0.9em;"
    );

    // Log response summary
    console.log("%cResponse Summary:", "font-weight: bold;");
    console.table({
      Status: `${response.status} ${response.statusText}`,
      Duration: `${duration}ms`,
      Timestamp: new Date(response.timestamp).toISOString(),
    });

    // Log response headers
    console.log("%cResponse Headers:", "font-weight: bold; color: #9C27B0;");
    console.table(response.headers);

    // Log response body
    if (response.body) {
      console.log("%cResponse Body:", "font-weight: bold; color: #4CAF50;");
      console.log(response.body);
    }

    console.groupEnd();
  }

  /**
   * Log an API error
   */
  logError(requestId: string, error: LoggedError): void {
    if (!this.enabled) return;

    const startTime = this.pendingRequests.get(requestId);
    const duration = startTime ? error.timestamp - startTime : error.duration;
    this.pendingRequests.delete(requestId);

    const style = "color: #F44336; font-weight: bold;";
    const resetStyle = "color: inherit; font-weight: normal;";

    console.groupCollapsed(
      `%c✖ API ERROR %c${error.status || "Network Error"} %c(${duration}ms)`,
      style,
      resetStyle,
      "color: #757575; font-size: 0.9em;"
    );

    // Log error summary
    console.log("%cError Details:", "font-weight: bold; color: #F44336;");
    console.table({
      Message: error.message,
      Status: error.status || "N/A",
      Duration: `${duration}ms`,
      Timestamp: new Date(error.timestamp).toISOString(),
    });

    // Log error details
    if (error.details) {
      console.log("%cError Details:", "font-weight: bold; color: #FF5722;");
      console.log(error.details);
    }

    // Log stack trace if available
    if (error.details instanceof Error) {
      console.log("%cStack Trace:", "font-weight: bold; color: #FF9800;");
      console.error(error.details);
    }

    console.groupEnd();
  }

  /**
   * Sanitize sensitive headers before logging
   */
  private sanitizeHeaders(headers: Record<string, string>): Record<string, string> {
    const sanitized: Record<string, string> = {};
    const sensitiveKeys = ["authorization", "cookie", "set-cookie", "x-api-key"];

    for (const [key, value] of Object.entries(headers)) {
      const lowerKey = key.toLowerCase();

      if (sensitiveKeys.includes(lowerKey)) {
        // Show partial value for debugging
        if (lowerKey === "authorization" && value.startsWith("Bearer ")) {
          const token = value.slice(7);
          sanitized[key] = `Bearer ${token.slice(0, 20)}...${token.slice(-20)}`;
        } else if (lowerKey === "cookie") {
          sanitized[key] = "[REDACTED - Contains session cookies]";
        } else {
          sanitized[key] = "[REDACTED]";
        }
      } else {
        sanitized[key] = value;
      }
    }

    return sanitized;
  }

  /**
   * Generate a unique request ID
   */
  generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
  }
}

// Singleton instance
export const apiLogger = new ApiLogger();

// Utility functions for common logging patterns
export const enableApiLogging = () => apiLogger.setEnabled(true);
export const disableApiLogging = () => apiLogger.setEnabled(false);

// Export for window access (debugging)
if (typeof window !== "undefined") {
  (window as any).__apiLogger = {
    enable: enableApiLogging,
    disable: disableApiLogging,
    instance: apiLogger,
  };
}

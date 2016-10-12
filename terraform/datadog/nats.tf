
resource "datadog_monitor" "nats" {
  name = "${format("%s NATS hosts", var.env)}"
  type = "service check"
  message = "Missing nats hosts in environment {{host.environment}}. Notify: @the-multi-cloud-paas-team@digital.cabinet-office.gov.uk"
  escalation_message = "Missing nats hosts! Check VM state."
  no_data_timeframe = "2"
  query = "${format("'datadog.agent.up'.over('environment:%s','job:nats').by('*').last(1).pct_by_status()", var.env)}"

  thresholds {
    ok = 0
    warning = 0
    critical = 10
  }

  require_full_window = true
  tags {
    "environment" = "${var.env}"
    "job" = "nats"
  }
}

resource "datadog_monitor" "nats_process_running" {
  name = "${format("%s NATS process running", var.env)}"
  type = "service check"
  message = "nats process not running. Check nats state."
  escalation_message = "nats process still not running. Check nats state."
  no_data_timeframe = "5"
  require_full_window = true

  query = "${format("'process.up'.over('environment:%s','process:nats').last(4).count_by_status()", var.env)}"

  thresholds {
    ok = 1
    warning = 2
    critical = 3
  }

  tags {
    "environment" = "${var.env}"
    "job" = "nats"
  }
}

resource "datadog_monitor" "nats_stream_forwarded_process_running" {
  name = "${format("%s NATS stream forwarder process running", var.env)}"
  type = "service check"
  message = "NATS stream forwarder process not running. Check nats state."
  escalation_message = "NATS stream forwarder process still not running. Check NATS state."
  no_data_timeframe = "5"
  require_full_window = true

  query = "${format("'process.up'.over('environment:%s','process:nats_stream_forwarder').last(4).count_by_status()", var.env)}"

  thresholds {
    ok = 1
    warning = 2
    critical = 3
  }

  tags {
    "environment" = "${var.env}"
    "job" = "nats"
  }
}

resource "datadog_monitor" "nats_service_open" {
  name = "${format("%s NATS service is accepting connections", var.env)}"
  type = "service check"
  message = "Large portion of NATS service are not accepting connections. Check deployment state."
  escalation_message = "Large portion of NATS service are still not accepting connections. Check deployment state."
  no_data_timeframe = "5"
  require_full_window = true

  query = "${format("'tcp.can_connect'.over('environment:%s','instance:nats_server').by('*').last(1).pct_by_status()", var.env)}"

  thresholds {
    critical = 50
  }

  tags {
    "environment" = "${var.env}"
    "job" = "router"
  }
}

resource "datadog_monitor" "nats_cluster_service_open" {
  name = "${format("%s NATS cluster service is accepting connections", var.env)}"
  type = "service check"
  message = "Large portion of NATS cluster service are not accepting connections. Check deployment state."
  escalation_message = "Large portion of NATS cluster service are still not accepting connections. Check deployment state."
  no_data_timeframe = "5"
  require_full_window = true

  query = "${format("'tcp.can_connect'.over('environment:%s','instance:nats_cluster').by('*').last(1).pct_by_status()", var.env)}"

  thresholds {
    critical = 50
  }

  tags {
    "environment" = "${var.env}"
    "job" = "router"
  }
}
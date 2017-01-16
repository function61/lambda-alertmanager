Use case: HTTP monitoring
=========================

You have an important web property that you want to monitor:

![](usecase_http-monitoring-page-screenshot.png)

You have AlertManager-Canary installed and configured to monitor it:

![](usecase_http-monitoring-canary-checkdefinition.png)

So, if Canary fails to find this text from the page:

```
Hostname: c70e24a08b3a
```

It'll send an alert (to configurable receivers), for example by SMS:

![](usecase_http-monitoring-sms.png)

And email:

![](usecase_http-monitoring-email.png)

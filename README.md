# Route service for Register on PaaS

One job, report to Google Analytics on every request.

## Requirements

`GOOGLE_ANALYTICS_TRACKING_ID` environment variable will need to be defined in order for the metrics to be stored approprietly. It has a form of: `UA-00000000-0`. Should be stored in the password store.

## How to deploy

This app can be hosted anywhere under the condition it has a public endpoint accessible.

There is no reason not to host this app on GOV.UK PaaS. Ideally, it should live in the `prod` namespace.

Manual deployment:

```sh
cf push
```

## How to use as route service

Once the application is hosted anywhere, we can setup CloudFoundry to route traffic through the app.

The following command will setup the configuration for the route service.

```sh
cf create-user-provided-service openregister-ga-route-service -r https://openregister-ga-route-service.cloudapps.digital
```

The following command will use the above configuration when the traffic comes to the application.

```sh
cf bind-route-service cloudapps.digital openregister-ga-route-service --hostname beta-multi
```

Assuming everything went smoothly, the service should be in play and used for routing.

## Troubleshoot

**Is this app running?**

Check the status of this app by running:

```sh
cf app google-analytics-reporter
```

**Is my service configured correctly?**

Check if initial config is present.

```sh
cf service openregister-ga-route-service
```

Check if the routes are attached to the application.

```sh
cf routes
```

**Still stuck?**

There are people working hard on [GOV.UK PaaS documentation](https://docs.cloud.service.gov.uk/deploying_services/route_services/), to deliver the best user experience possible. See if perhaps that documentation covers your problem.

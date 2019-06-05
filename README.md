# GOV.UK Registers route service

GOV.UK Registers collects usage data via Google Analytics. The route service is responsible for proxying traffic to the Registers application and sending data based on user requests to Google Analytics.

## Requirements

The `GOOGLE_ANALYTICS_TRACKING_ID` environment variable will need to be defined in order for the metrics to be stored appropriately. It takes the form `UA-00000000-0` and should be stored in the password store.

## How to deploy

This route service is automatically deployed to GOV.UK PaaS. However, to manually deploy it to GOV.UK PaaS:

```sh
cf push
```

## Enable the route service

Once the route service is deployed, we can tell the PaaS to route traffic through the service.

The following command will setup the configuration for the route service.

```sh
cf create-user-provided-service openregister-ga-route-service -r https://openregister-ga-route-service.cloudapps.digital
```

The following command will enable the route service to proxy traffic to the destination registers application.

```sh
cf bind-route-service cloudapps.digital openregister-ga-route-service --hostname <name of destination app>
```

Assuming everything went smoothly, the service should be in play and used for routing.

## Troubleshoot

### Configuration

Debug-level logging can be enabled for the route service by setting a `DEBUG` environment variable with a value of `true`. The Google Analytics URL can be overridden with the `GOOGLE_ANALYTICS_URL` environment variable.

### Is the route service running?

Check the status of this app by running:

```sh
cf app google-analytics-reporter
```

### Is the route service configured correctly?

Check if initial config is present.

```sh
cf service openregister-ga-route-service
```

Check if the correct routes are attached to the application.

```sh
cf routes
```

### Still stuck?

There are people working hard on [GOV.UK PaaS documentation](https://docs.cloud.service.gov.uk/deploying_services/route_services/), to deliver the best user experience possible. See if perhaps that documentation covers your problem.

# AWS Console Log Aggregator
> A tool for aggregating aws instances console log

## Basic Usage Example
### Aggregating 2 AWS instances console log on region eu-central-1
##### aggregated logs locations: /var/log/InstanceId1.log, /var/log/InstanceId2.log
    ./aggregator -region eu-central-1 -folder /var/log -id instanceId1 -id InstanceId2

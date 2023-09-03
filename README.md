# LCM - Local Cloud Metadata


## Main Account

Will use ApplicationDefaultCredentials by default.

Make sure the main account has `roles/iam.serviceAccountTokenCreator` on the project, which will propagate to service accounts, or that it has the correct privileges to grant itself the token creator role on requested service account on demand.

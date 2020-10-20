#!/bin/bash
# Environment- specific variables
# TODO: automatize with Terraform and Vault
. scripts/cluster.env

export db_port="5432"
export db_host="172.23.32.3"

export db_user_auth=service-auth
export db_pass_auth=aesohc4Ahth6ahsh
export db_name_auth=service-auth

export db_user_bill=service-bill
export db_pass_bill=thah7Aijooquo3Ze
export db_name_bill=service-bill

export db_user_customer=service-customer
export db_pass_customer=aeBeex4Chaitieci
export db_name_customer=service-customer

export db_user_employee=service-employee
export db_pass_employee=Xaef4faeGiwophai
export db_name_employee=service-employee

export db_user_merchant=service-merchant
export db_pass_merchant=Kaip0phooTangiet
export db_name_merchant=service-merchant

export db_user_notification=service-notification
export db_pass_notification=Fahbeix2quaegah3
export db_name_notification=service-notification

export db_user_purchase=service-purchase
export db_pass_purchase=Vei8beiyooN1lahz
export db_name_purchase=service-purchase

export db_user_receipt=service-receipt
export db_pass_receipt=ke9tai4ooM2efieh
export db_name_receipt=service-receipt

export db_user_recovery=service-recovery
export db_pass_recovery=sah1Beed5jieLahg
export db_name_recovery=service-recovery

export db_user_wallet=service-wallet
export db_pass_wallet=quesiej3Phaiquai
export db_name_wallet=service-wallet

export db_user_integration=service-integration
export db_pass_integration=Bu1ooweimae9iLae
export db_name_integration=service-integration

export db_user_fraud=service-fraud
export db_pass_fraud=oa01Uj0MkL3s91q2Ol00oP
export db_name_fraud=service-fraud

# Services
export service_wallet_host="http://service-wallet-exp01:37007"
export service_auth_host="http://service-auth-exp01:37001"
export service_customer_host="http://service-customer-exp01:37002"
export service_purchase_host="http://service-purchase-exp01:37006"
export service_merchant_host="http://service-merchant-exp01:37008"


# ES
export elasticsearch_host="e93effcb1ecd4b9fa678a1ddcfa148a0.europe-west1.gcp.cloud.es.io"
export elasticsearch_port="9243"
export elasticsearch_username="elastic"
export elasticsearch_password="3kjBOlteI1ldxrahh5581IPC"

# Redis
export redis_host="redis-13602.c102.us-east-1-mz.ec2.cloud.redislabs.com"
export redis_port="13602"
export redis_pass="NLSj5mxET9najXYL7eTHfm1z3cRiaMCA"

# Rabbit
export rabbit_host="bulldog.rmq.cloudamqp.com"
export rabbit_port="5672"
export rabbit_ssl_port="443"
export rabbit_user="kcugblqu"
export rabbit_pass="5IgM6MwRL_tNFXjuYTS7AjzavesFQQVr"
export rabbit_vhost="kcugblqu"

# MongoDB
export mongodb_uri="mongodb+srv://service-recovery-exp01:GOLbxcXu6K6t8mNk@cluster0-fedpm.gcp.mongodb.net/test?retryWrites=true&w=majority"
export mongodb_uri_integration="mongodb+srv://service-integration-exp01:eAVvkQVIsbVCXBtY@cluster0-fedpm.gcp.mongodb.net/test?retryWrites=true&w=majority"

# Enable bitbucket photos processing
export gs_user="receipts-storage@development-gg.iam.gserviceaccount.com"
export gs_key="-----BEGIN PRIVATE KEY-----MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC0UoQtQTBHIaSHjGigaadFnCXuG4cV4qO70d3ZdPzz5AieXt76OXyPbLdlY9x8vmlEv2sVmYtIBY1zuXxRBuVVgzbJlhUm0M3Tvp4LC5xmFir6sS4MOKRhx3gEZzXqESQ7T6VIl4dqeFUamMsbsTdGjHA+vW1JJgzgysYEgO06/efM4sU2H2OhKrcyJ1U0bxkj90qVAKPQmwt4NGm4LYkXDGLOe7PZN6o3GMMTj+N+UJmghsPHcC/arEpfbjSCGTjTlKKhcJQqa5obzrtzwX0wobpgXFia972aJCQvjZb/zHbGuLj4K3YE0sZeLg7Wy5LEvF70rttUpYYgP03SMt9bAgMBAAECggEAB7PSoFnXGK+JZwTXX0Rdq2DtID7RLJBYYBup0OUcMCrLndnKYOcQ7AUMnGpOLvUcBNGMXtYGytYtFNL8Fwkm+Ka4y2Muyh0Qynx4JPUJz2gfZw/0bWJqAYc5fpUOCrTqvRUJp8OMbdhwAi/VCGMvskLDdKdqW9dVPrDkRNyUK/kKjo1h8e+k95WOPm5pmx5bskmB835L6G/0CrscvWPZOIPu3RuEXYUmKkOfkG/q23OEymtI8LqOGwYwH4yZPoZq5X/xjwXsN9XAFKT1UGhziwCot72cMQvwmxLbFkEgJrfu3SIqHoyWAgqNs7JRu9lnVqKgsndbpFIUn4BnHrrymQKBgQDedRHfDdfFsf220VCI7eTslcRhX+IO2xLj/RNDmmr625BGIqBeo0pDokEc6VC05IOA9qLZ3f8uOSKeAuOsWrnsBJIV09zJgWl3aVNUmX8tAmp5W9M32C6TNsqTBxlTzXnAsjA1op77qcsRY3MxQH8BcJE28QgaTVvGEqN7X6uJTwKBgQDPgwWLxwQ1zwxHXI3bAUe5933ueY4OnAj6qeZ4IsprTW5g3qFr1N9BSaLvjkicYMa6S3jnn9p/coGvfqiQZB/xWYgv4F8hy75FOn+ct/gZLb14eM84VL5L93QlPTY0iuRfG1748cTnx6yOflJpI8/MBEOm/ZesaKtP8HiOw3juNQKBgQCLPMNe1Y9Ekk+3afP6gMxUuLkeKaGYos6EHRc9rR1gvqTjATFXiuUkyB3xNqfpUU5uHfF4ZFcgW2qrdCuE6ZSNgZ7eQqljBrk4oJgjz5+mUGjMZQkjXxBn3FeXB053AZk/X0iFia/w3SnZTGIBZdkY0ZhSxzLHI7xZkbj5s7vuSQKBgHn/q8kLznvcKHnj/jpdvE+nI9CKgmwwbE8CiE7lFWCUe2pUOU7uLftyUWrJmgLmGq/4IzL6FjmLlpcYvf12ABmi66BKJ2P1Jv4IcHIw7pnO/G/RhvK1T9PVveEO5clqRu1raCCv83XZPKfhuI270jU95JBO01c3ilBLLnWwkm5pAoGBAMcfTuZ3HdNc4UORFAPvjQpJMtj5v8iC5nJrEo7OO8V52/v4VVaqXTjXHmlHVyRpRxo7xPy+RWPbLL/q8myTM4NZ6ZBW1fKo4h/EzPVgBeXaCaH4t+pjflLoG2OpP43XPPDStGzs9kpmMnhQhgC0K0qOoUi1gVzn6JJM2YU7Kkus-----END PRIVATE KEY-----"
export gs_base_url="https://storage.googleapis.com"
export gs_bucket_path="/zipzero-receipts/"

# Service bill credentials
export service_bill_login="service-bill"
export service_bill_pass="abi*2(ksdfK2"

# Service purchase credentials
export service_purchase_login="service-purchase"
export service_purchase_pass="Ola&2!@#kK9;"

# Service receipt credentials
export service_receipt_login="service-receipt"
export service_receipt_pass="Ikw80*^;2a"

# Service search credentials
export service_search_login="service-search"
export service_search_pass="Ymz80*^;5a"

# Flags used to check customer bank account (if its corporate or not)
export sort_code_validation=true

# Api code for sort code service
export sort_codes_api_key="2e44acb64ce5b7ce4e5752d16122cb4d"

# Secret used to check JWT
export jwt_secret="jwt_secret_key_JFIWiifs8823kKf28k"

# Token used to generate JWT
export token_secret="token_secret_key_jKii2kLillKKilsa82"

# Flag used to check user mail addresess (if they're one "some" list)
export customer_list_enabled=false

# Application public URL
export app_host_url="https://app.zipzero.com"
export app_host_url_crm="https://crm-exp01.zipzero.com"

# Some email addressess used for.. ?
export default_admin_address="help@zipzero.com"
export default_reply_to_address="help@zipzero.com"
export default_from_address="help@zipzero.com"

# Days limit, after which customer can't scan the bill
export scan_receipt_limit_in_days=7

# API address and key for micro blik service
export micro_blink_api_url="https://scan.blinkreceipt.com/api_scan/v12"
export micro_blink_api_key="87c8d5768371464e8beefbad3d9981cb"
export mb_receipt_validation=false
export zz_receipt_validation=false

# Statistics mail send schedule and address
export statistic_email_time="0 0 9 * * *"
export statistics_address="stats.exptwo@zipzero.com"

# Offers URL
export my_offers_url="https://offers-exp01.zipzero.com?webLoginToken="

#Integrations
export environment_name="TST01"
export trade_doubler_access_token="8f728369-ccb6-4b86-a924-58cf766184c7"
export partnerize_publisher_id="1101l90375"
export partnerize_user_api_key="6ZIXurgM"
export partnerize_user_application_key="Pto8eGaqX8"
export trade_tracker_account="206658"
export trade_tracker_passphrase="097db06609143e7228040de11ba1fc34cd2a51b9"
export trade_tracker_site="356214"
export awin_access_token="2ce7ad15-52f7-46e1-bda7-ac4165f2fcab"
export awin_publisher_id="629753"
export cj_affiliate_scheduling_enabled=false
export partnerize_scheduling_enabled=false
export awin_scheduling_enabled=false
export trade_tracker_scheduling_enabled=false
export partnerize_fixed_delay_in_ms=43200000
export partnerize_fixed_delay_in_ms_latest=120000
export awin_fixed_delay_in_ms=43200000
export awin_fixed_delay_in_ms_latest=120000
export trade_tracker_fixed_delay_in_ms=43200000
export trade_tracker_fixed_delay_in_ms_latest=120000
export cj_affiliate_fixed_delay_in_ms=43200000
export cj_affiliate_fixed_delay_in_ms_latest=120000
export token_expiration_millis=10800000
export zz_bank_account="309897-76492668"
export local_zone_id="UTC"

# CRM URL
export api_endpoint_url="https://api-exp01.zipzero.com"
export envid="exp01"

#Beta testers
export beta_tester_access_email_enabled=true

# Frauds
export frauds_script_run_time="0 0 3 * * *"
export service_fraud_login="service-fraud"
export service_fraud_pass="HKk80*f..12b"

# Landing Page
export landing_page_branch="grzeg-test1"

# keymaster

![Go](https://github.com/bsycorp/keymaster/workflows/Go/badge.svg?branch=master)

Single sign-on for SSH, AWS IAM &amp; Kubernetes, for humans and machines.

See: [docs/deployment.md](docs/deployment.md)

And: [terraform/aws/issuing_lambda/](terraform/aws/issuing_lambda/)

## Client Usage examples

### Generic CI

For ci, you need to specify the usual issuer and target role, 
as well as details to identify the source of your access request.

```
km ci --issuer <issuing-lambda> --role deployment \
  --username smithb12 \
  --name "Bob Smith" \
  --email "bob.smith@awesome.com" \
  --description "enhance the magic" \
  --url "https://github.com/bsycorp/keymaster/pull/7"
```

All fields are required in ci usage.

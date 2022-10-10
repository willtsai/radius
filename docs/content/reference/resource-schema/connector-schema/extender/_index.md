---
type: docs
title: "Extender resource"
linkTitle: "Extender"
description: "Learn how to use Extender resource in Radius"
weight: 999
slug: "extender"
---

## Overview

An extender resource could be used to bring in a custom resource into Radius for which there is no first class support to "extend" the Radius functionality. The resource can define arbitrary key-value pairs and secrets. These properties and secret values can then be used to connect it to other Radius resources.

## Resource format

{{< rad file="snippets/extender.bicep" embed=true marker="//EXTENDER" >}}

### Top-level

| Key  | Required | Description | Example |
|------|:--------:|-------------|---------|
| name | y | The name of your resource. | `mongo`
| location | y | The location of your resource. See [common values]({{< ref "resource-schema.md#common-values" >}}) for more information. | `global`
| [properties](#properties) | y | Properties of the resource. | [See below](#properties)

### Properties

| Key  | Required | Description | Example |
|------|:--------:|-------------|---------|
| \<user-defined key-value pairs\> | n | User-defined properties of the extender. Can accept any key name except 'secrets'. | `fromNumber: '222-222-2222'`
| secrets | n | Secrets in the form of key-value pairs | `password: '******'`

## Methods

The following methods are available on the Redis connector:

| Method | Description |
|--------|-------------|
| secrets('SECRET_NAME') | Get the value of a secret. |
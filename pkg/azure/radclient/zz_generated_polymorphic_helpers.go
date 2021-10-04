// +build go1.13

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package radclient

import "encoding/json"

func unmarshalComponentTraitClassification(rawMsg json.RawMessage) (ComponentTraitClassification, error) {
	if rawMsg == nil {
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(rawMsg, &m); err != nil {
		return nil, err
	}
	var b ComponentTraitClassification
	switch m["kind"] {
	case "dapr.io/Sidecar@v1alpha1":
		b = &DaprTrait{}
	case "radius.dev/InboundRoute@v1alpha1":
		b = &InboundRouteTrait{}
	case "radius.dev/ManualScaling@v1alpha1":
		b = &ManualScalingTrait{}
	default:
		b = &ComponentTrait{}
	}
	return b, json.Unmarshal(rawMsg, b)
}

func unmarshalComponentTraitClassificationArray(rawMsg json.RawMessage) ([]ComponentTraitClassification, error) {
	if rawMsg == nil {
		return nil, nil
	}
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(rawMsg, &rawMessages); err != nil {
		return nil, err
	}
	fArray := make([]ComponentTraitClassification, len(rawMessages))
	for index, rawMessage := range rawMessages {
		f, err := unmarshalComponentTraitClassification(rawMessage)
		if err != nil {
			return nil, err
		}
		fArray[index] = f
	}
	return fArray, nil
}


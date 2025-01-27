// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

// Package model
//
// What-changed models are unified across OpenAPI and Swagger. Everything is kept flat for simplicity, so please
// excuse the size of the package. There is a lot of data to crunch!
//
// Every model in here is either universal (works across both versions of OpenAPI) or is bound to a specific version
// of OpenAPI. There is only a single model however - so version specific objects are marked accordingly.
package model

import (
    "github.com/pb33f/libopenapi/datamodel/low"
    "github.com/pb33f/libopenapi/datamodel/low/base"
    "github.com/pb33f/libopenapi/datamodel/low/v2"
    "github.com/pb33f/libopenapi/datamodel/low/v3"
    "reflect"
)

// DocumentChanges represents all the changes made to an OpenAPI document.
type DocumentChanges struct {
    *PropertyChanges
    InfoChanges                *InfoChanges                  `json:"info,omitempty" yaml:"info,omitempty"`
    PathsChanges               *PathsChanges                 `json:"paths,omitempty" yaml:"paths,omitempty"`
    TagChanges                 []*TagChanges                 `json:"tags,omitempty" yaml:"tags,omitempty"`
    ExternalDocChanges         *ExternalDocChanges           `json:"externalDoc,omitempty" yaml:"externalDoc,omitempty"`
    WebhookChanges             map[string]*PathItemChanges   `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
    ServerChanges              []*ServerChanges              `json:"servers,omitempty" yaml:"servers,omitempty"`
    SecurityRequirementChanges []*SecurityRequirementChanges `json:"securityRequirements,omitempty" yaml:"securityRequirements,omitempty"`
    ComponentsChanges          *ComponentsChanges            `json:"components,omitempty" yaml:"components,omitempty"`
    ExtensionChanges           *ExtensionChanges             `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

// TotalChanges returns a total count of all changes made in the Document
func (d *DocumentChanges) TotalChanges() int {
    c := d.PropertyChanges.TotalChanges()
    if d.InfoChanges != nil {
        c += d.InfoChanges.TotalChanges()
    }
    if d.PathsChanges != nil {
        c += d.PathsChanges.TotalChanges()
    }
    for k := range d.TagChanges {
        c += d.TagChanges[k].TotalChanges()
    }
    if d.ExternalDocChanges != nil {
        c += d.ExternalDocChanges.TotalChanges()
    }
    for k := range d.WebhookChanges {
        c += d.WebhookChanges[k].TotalChanges()
    }
    for k := range d.ServerChanges {
        c += d.ServerChanges[k].TotalChanges()
    }
    for k := range d.SecurityRequirementChanges {
        c += d.SecurityRequirementChanges[k].TotalChanges()
    }
    if d.ComponentsChanges != nil {
        c += d.ComponentsChanges.TotalChanges()
    }
    if d.ExtensionChanges != nil {
        c += d.ExtensionChanges.TotalChanges()
    }
    return c
}

// TotalBreakingChanges returns a total count of all breaking changes made in the Document
func (d *DocumentChanges) TotalBreakingChanges() int {
    c := d.PropertyChanges.TotalBreakingChanges()
    if d.InfoChanges != nil {
        c += d.InfoChanges.TotalBreakingChanges()
    }
    if d.PathsChanges != nil {
        c += d.PathsChanges.TotalBreakingChanges()
    }
    for k := range d.TagChanges {
        c += d.TagChanges[k].TotalBreakingChanges()
    }
    if d.ExternalDocChanges != nil {
        c += d.ExternalDocChanges.TotalBreakingChanges()
    }
    for k := range d.WebhookChanges {
        c += d.WebhookChanges[k].TotalBreakingChanges()
    }
    for k := range d.ServerChanges {
        c += d.ServerChanges[k].TotalBreakingChanges()
    }
    for k := range d.SecurityRequirementChanges {
        c += d.SecurityRequirementChanges[k].TotalBreakingChanges()
    }
    if d.ComponentsChanges != nil {
        c += d.ComponentsChanges.TotalBreakingChanges()
    }
    return c
}

// CompareDocuments will compare any two OpenAPI documents (either Swagger or OpenAPI) and return a pointer to
// DocumentChanges that outlines everything that was found to have changed.
func CompareDocuments(l, r any) *DocumentChanges {

    var changes []*Change
    var props []*PropertyCheck

    dc := new(DocumentChanges)

    if reflect.TypeOf(&v2.Swagger{}) == reflect.TypeOf(l) && reflect.TypeOf(&v2.Swagger{}) == reflect.TypeOf(r) {
        lDoc := l.(*v2.Swagger)
        rDoc := r.(*v2.Swagger)

        // version
        addPropertyCheck(&props, lDoc.Swagger.ValueNode, rDoc.Swagger.ValueNode,
            lDoc.Swagger.Value, rDoc.Swagger.Value, &changes, v3.SwaggerLabel, true)

        // host
        addPropertyCheck(&props, lDoc.Host.ValueNode, rDoc.Host.ValueNode,
            lDoc.Host.Value, rDoc.Host.Value, &changes, v3.HostLabel, true)

        // base path
        addPropertyCheck(&props, lDoc.BasePath.ValueNode, rDoc.BasePath.ValueNode,
            lDoc.BasePath.Value, rDoc.BasePath.Value, &changes, v3.BasePathLabel, true)

        // schemes
        if len(lDoc.Schemes.Value) > 0 || len(lDoc.Schemes.Value) > 0 {
            ExtractStringValueSliceChanges(lDoc.Schemes.Value, rDoc.Schemes.Value,
                &changes, v3.SchemesLabel, true)
        }
        // consumes
        if len(lDoc.Consumes.Value) > 0 || len(lDoc.Consumes.Value) > 0 {
            ExtractStringValueSliceChanges(lDoc.Consumes.Value, rDoc.Consumes.Value,
                &changes, v3.ConsumesLabel, true)
        }
        // produces
        if len(lDoc.Produces.Value) > 0 || len(lDoc.Produces.Value) > 0 {
            ExtractStringValueSliceChanges(lDoc.Produces.Value, rDoc.Produces.Value,
                &changes, v3.ProducesLabel, true)
        }

        // tags
        dc.TagChanges = CompareTags(lDoc.Tags.Value, rDoc.Tags.Value)

        // paths
        if !lDoc.Paths.IsEmpty() || !rDoc.Paths.IsEmpty() {
            dc.PathsChanges = ComparePaths(lDoc.Paths.Value, rDoc.Paths.Value)
        }

        // external docs
        compareDocumentExternalDocs(lDoc, rDoc, dc, &changes)

        // info
        compareDocumentInfo(&lDoc.Info, &rDoc.Info, dc, &changes)

        // security
        if !lDoc.Security.IsEmpty() || !rDoc.Security.IsEmpty() {
            checkSecurity(lDoc.Security, rDoc.Security, &changes, dc)
        }

        // components / definitions
        // swagger (damn you) decided to put all this stuff at the document root, rather than cleanly
        // placing it under a parent, like they did with OpenAPI. This means picking through each definition
        // creating a new set of changes and then morphing them into a single changes object.
        cc := new(ComponentsChanges)
        cc.PropertyChanges = new(PropertyChanges)
        if n := CompareComponents(lDoc.Definitions.Value, rDoc.Definitions.Value); n != nil {
            cc.SchemaChanges = n.SchemaChanges
        }
        if n := CompareComponents(lDoc.SecurityDefinitions.Value, rDoc.SecurityDefinitions.Value); n != nil {
            cc.SecuritySchemeChanges = n.SecuritySchemeChanges
        }
        if n := CompareComponents(lDoc.Parameters.Value, rDoc.Parameters.Value); n != nil {
            cc.PropertyChanges.Changes = append(cc.PropertyChanges.Changes, n.Changes...)
        }
        if n := CompareComponents(lDoc.Responses.Value, rDoc.Responses.Value); n != nil {
            cc.Changes = append(cc.Changes, n.Changes...)
        }
        dc.ExtensionChanges = CompareExtensions(lDoc.Extensions, rDoc.Extensions)
        if cc.TotalChanges() > 0 {
            dc.ComponentsChanges = cc
        }
    }

    if reflect.TypeOf(&v3.Document{}) == reflect.TypeOf(l) && reflect.TypeOf(&v3.Document{}) == reflect.TypeOf(r) {
        lDoc := l.(*v3.Document)
        rDoc := r.(*v3.Document)

        // version
        addPropertyCheck(&props, lDoc.Version.ValueNode, rDoc.Version.ValueNode,
            lDoc.Version.Value, rDoc.Version.Value, &changes, v3.OpenAPILabel, true)

        // schema dialect
        addPropertyCheck(&props, lDoc.JsonSchemaDialect.ValueNode, rDoc.JsonSchemaDialect.ValueNode,
            lDoc.JsonSchemaDialect.Value, rDoc.JsonSchemaDialect.Value, &changes, v3.JSONSchemaDialectLabel, true)

        // tags
        dc.TagChanges = CompareTags(lDoc.Tags.Value, rDoc.Tags.Value)

        // paths
        if !lDoc.Paths.IsEmpty() || !rDoc.Paths.IsEmpty() {
            dc.PathsChanges = ComparePaths(lDoc.Paths.Value, rDoc.Paths.Value)
        }

        // external docs
        compareDocumentExternalDocs(lDoc, rDoc, dc, &changes)

        // info
        compareDocumentInfo(&lDoc.Info, &rDoc.Info, dc, &changes)

        // security
        if !lDoc.Security.IsEmpty() || !rDoc.Security.IsEmpty() {
            checkSecurity(lDoc.Security, rDoc.Security, &changes, dc)
        }

        // compare components.
        if !lDoc.Components.IsEmpty() && !rDoc.Components.IsEmpty() {
            if n := CompareComponents(lDoc.Components.Value, rDoc.Components.Value); n != nil {
                dc.ComponentsChanges = n
            }
        }
        if !lDoc.Components.IsEmpty() && rDoc.Components.IsEmpty() {
            CreateChange(&changes, PropertyRemoved, v3.ComponentsLabel,
                lDoc.Components.ValueNode, nil, true, lDoc.Components.Value, nil)
        }
        if lDoc.Components.IsEmpty() && !rDoc.Components.IsEmpty() {
            CreateChange(&changes, PropertyAdded, v3.ComponentsLabel,
                rDoc.Components.ValueNode, nil, false, nil, lDoc.Components.Value)
        }

        // compare servers
        if n := checkServers(lDoc.Servers, rDoc.Servers); n != nil {
            dc.ServerChanges = n
        }

        // compare webhooks
        dc.WebhookChanges = CheckMapForChanges(lDoc.Webhooks.Value, rDoc.Webhooks.Value, &changes,
            v3.WebhooksLabel, ComparePathItemsV3)

        // extensions
        dc.ExtensionChanges = CompareExtensions(lDoc.Extensions, rDoc.Extensions)
    }

    CheckProperties(props)
    dc.PropertyChanges = NewPropertyChanges(changes)
    if dc.TotalChanges() <= 0 {
        return nil
    }
    return dc
}

func compareDocumentExternalDocs(l, r low.HasExternalDocs, dc *DocumentChanges, changes *[]*Change) {
    // external docs
    if !l.GetExternalDocs().IsEmpty() && !r.GetExternalDocs().IsEmpty() {
        lExtDoc := l.GetExternalDocs().Value.(*base.ExternalDoc)
        rExtDoc := r.GetExternalDocs().Value.(*base.ExternalDoc)
        if !low.AreEqual(lExtDoc, rExtDoc) {
            dc.ExternalDocChanges = CompareExternalDocs(lExtDoc, rExtDoc)
        }
    }
    if l.GetExternalDocs().IsEmpty() && !r.GetExternalDocs().IsEmpty() {
        CreateChange(changes, PropertyAdded, v3.ExternalDocsLabel,
            nil, r.GetExternalDocs().ValueNode, false, nil,
            r.GetExternalDocs().Value)
    }
    if !l.GetExternalDocs().IsEmpty() && r.GetExternalDocs().IsEmpty() {
        CreateChange(changes, PropertyRemoved, v3.ExternalDocsLabel,
            l.GetExternalDocs().ValueNode, nil, false, l.GetExternalDocs().Value,
            nil)
    }
}

func compareDocumentInfo(l, r *low.NodeReference[*base.Info], dc *DocumentChanges, changes *[]*Change) {
    // info
    if !l.IsEmpty() && !r.IsEmpty() {
        lInfo := l.Value
        rInfo := r.Value
        if !low.AreEqual(lInfo, rInfo) {
            dc.InfoChanges = CompareInfo(lInfo, rInfo)
        }
    }
    if l.IsEmpty() && !r.IsEmpty() {
        CreateChange(changes, PropertyAdded, v3.InfoLabel,
            nil, r.ValueNode, false, nil,
            r.Value)
    }
    if !l.IsEmpty() && r.IsEmpty() {
        CreateChange(changes, PropertyRemoved, v3.InfoLabel,
            l.ValueNode, nil, false, l.Value,
            nil)
    }
}

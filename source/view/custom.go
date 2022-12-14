package view

import (
    "fmt"
    "reflect"
    "github.com/satori/go.uuid"
    //"errors"

    "woodchuck/model/bucket"
    "woodchuck/model/object"
    "woodchuck/instance/log"
    "woodchuck/instance/views"
)

//
type Value interface{
}

//
type Custom struct {
    Type string                         `json:"type"`
    Attributes map[string]Value         `json:"attributes"`
}

//
func NewCustom() (*Custom, error) {
    return &Custom{
        Attributes: make(map[string]Value),
    }, nil
}

// 
func NewCustomByBucket(bucket *bucket.Bucket) (*Custom, error) {
    log.Debug("start", "view.NewCustomByBucket: bucket", fmt.Sprintf("%+v", bucket))

    bService, err := bucket.GetService()
    if err != nil {
        log.Error("error", "view.NewCustomByBucket:", err)
        return nil, err
    }

    bType, err := bucket.GetType()
    if err != nil {
        log.Error("error", "view.NewCustomByBucket:", err)
        return nil, err
    }

    instance := views.GetInstance()
    view, err := instance.GetViewByType(bService, "Bucket", bType)
    if err != nil {
        log.Error("error", "view.NewCustomByBucket:", err)
        return nil, err
    }
    log.Debug("view", "view.NewCustomByBucket:", fmt.Sprintf("%+v", view))

    custom, err := NewCustom()
    if err != nil {
        log.Error("error", "view.NewCustomByBucket:", err)
        return nil, err
    }
    custom.Type = bType

    vBucket := reflect.ValueOf(bucket)
    vBucket = reflect.Indirect(vBucket)

    for _, attr := range view.Attributes {
        if( attr.Tags ) {
            value, _ := bucket.GetTagValueByKey(attr.Name)
            log.Debug("value", "view.NewCustomByBucket: attribute =", attr.Name, ", value =", value)
            custom.SetValueByAttributeName(attr.Name, value)
        } else {
            for i := 0; i < vBucket.NumField(); i++ {
                field := vBucket.Type().Field(i)
                name := field.Tag.Get("json")

                if attr.Name == name {
                    value := vBucket.FieldByName(field.Name)
                    if field.Type.Name() == "string" {
                        log.Debug("value", "view.NewCustomByBucket: attribute =", attr.Name, ", value =", value.String())
                        custom.SetValueByAttributeName(attr.Name, value.String())
                    } else if field.Type.Name() == "bool" {
                        log.Debug("value", "view.NewCustomByBucket: attribute =", attr.Name, ", value =", value.Bool())
                        custom.SetValueByAttributeName(attr.Name, value.Bool())
                    }
                    break
                }
            }
        }
    }

    log.Debug("success", "view.NewCustomByBucket:", fmt.Sprintf("%+v", custom))
    return custom, nil
}

//
func (c *Custom) GetValueByAttributeName(name string) Value {
    return c.Attributes[name]
}

//
func (c *Custom) SetValueByAttributeName(name string, value Value) error {
    c.Attributes[name] = value
    return nil
}

// 
func (c *Custom) NewCreateBucket(service string) (*bucket.CreateBucket, error) {
    log.Debug("start", "NewCreateBucket: service =", service)

    instance := views.GetInstance()
    view, err := instance.GetViewByType(service, "CreateBucket", c.Type)
    if err != nil {
        log.Error("error", "NewCreateBucket:", err)
        return nil, err
    }

    create, err := bucket.NewCreateBucket()
    if err != nil {
        log.Error("error", "NewCreateBucket:", err)
        return nil, err
    }

    create.SetService(service)
    create.SetType(c.Type)

    vcreate := reflect.ValueOf(create)
    vcreate = reflect.Indirect(vcreate)

    for _, attr := range view.Attributes {
        log.Debug("attribute", attr)

        var val reflect.Value
        if attr.Filled != nil {
            val = reflect.ValueOf(attr.Filled)
        } else {
            value := c.GetValueByAttributeName(attr.Name)
            val = reflect.ValueOf(value)
        }
        log.Debug("value", val)

        if attr.Tags {
            err = create.SetTag(attr.Name, val.String())
            if err != nil {
                break
            }
        } else {
            for i := 0; i < vcreate.NumField(); i++ {
                field := vcreate.Type().Field(i)
                name := field.Tag.Get("json")

                if attr.Name == name {
                    log.Debug("field", "attr.Name =", attr.Name, ", name =", name)

                    if field.Type.Name() == "string" {
                        str := val.String()

                        if str == "func.GetUUID()" {
                            str = fmt.Sprintf("%s", uuid.NewV4())
                        }

                        vcreate.FieldByName(field.Name).SetString(str)
                        log.Debug("filed", "SetString:", field.Name, "=>", str)
                    } else if field.Type.Name() == "bool" {
                        vcreate.FieldByName(field.Name).SetBool(val.Bool())
                        log.Debug("field", "SetBool:", field.Name, "=>", val.Bool())
                    }
                    break
                }
            }
        }
    }

    if err != nil {
        log.Error("error", "NewCreateBucket:", err)
        return nil, err
    }

    log.Debug("success", "NewCreateBucket: bucket =", fmt.Sprintf("%+v", create))
    return create, nil
}

//
func (c *Custom) NewCreateObject(service string, bucket_name string) (*object.CreateObject, error) {
    log.Debug("start", "NewCreateObject: service =", service)

    instance := views.GetInstance()
    view, err := instance.GetViewByType(service, "CreateObject", c.Type)
    if err != nil {
        log.Error("error", "NewCreateObject:", err)
        return nil, err
    }
    log.Debug("view", fmt.Sprintf("%+v", view))

    create, err := object.NewCreateObject()
    if err != nil {
        log.Error("error", "NewCreateObject:", err)
        return nil, err
    }

    create.SetService(service)
    create.SetType(c.Type)
    create.Bucket = bucket_name

    vcreate := reflect.ValueOf(create)
    vcreate = reflect.Indirect(vcreate)

    for _, attr := range view.Attributes {
        log.Debug("attribute", attr)

        var val reflect.Value
        if attr.Filled != nil {
            val = reflect.ValueOf(attr.Filled)
        } else {
            value := c.GetValueByAttributeName(attr.Name)
            val = reflect.ValueOf(value)
        }

        if !val.IsValid() {
            continue
        }
        log.Debug("value", val)

        if attr.Tags {
            err = create.SetTag(attr.Name, val.String())
            if err != nil {
                break
            }
        } else {
            for i := 0; i < vcreate.NumField(); i++ {
                field := vcreate.Type().Field(i)
                name := field.Tag.Get("json")

                if attr.Name == name {
                    log.Debug("field", "attr.Name =", attr.Name, ", name =", name)

                    if field.Type.Name() == "string" {
                        str := val.String()
                        vcreate.FieldByName(field.Name).SetString(str)
                        log.Debug("filed", "SetString:", field.Name, "=>", str)
                    } else if field.Type.Name() == "bool" {
                        vcreate.FieldByName(field.Name).SetBool(val.Bool())
                        log.Debug("field", "SetBool:", field.Name, "=>", val.Bool())
                    }
                    break
                }
            }
        }
    }

    log.Debug("success", "NewCreateObject: object =", fmt.Sprintf("%+v", create))
    return create, nil
}

//
func NewCustomByObject(object *object.Object) (*Custom, error) {
    log.Debug("start", "view.NewCustomByObject: object =", fmt.Sprintf("%+v", object))

    oService, err := object.GetService()
    if err != nil {
        log.Error("error", "view.NewCustomByObject:", err)
        return nil, err
    }

    oType, err := object.GetType()
    if err != nil {
        log.Error("error", "view.NewCustomByObject:", err)
        return nil, err
    }

    instance := views.GetInstance()
    view, err := instance.GetViewByType(oService, "Object", oType)
    if err != nil {
        log.Error("error", "view.NewCustomByObject:", err)
        return nil, err
    }
    log.Debug("view", "view.NewCustomByObject:", fmt.Sprintf("%+v", view))

    custom, err := NewCustom()
    if err != nil {
        log.Error("error", "view.NewCustomByObject:", err)
        return nil, err
    }
    custom.Type = oType

    vObject := reflect.ValueOf(object)
    vObject = reflect.Indirect(vObject)

    for _, attr := range view.Attributes {
        if( attr.Tags ) {
            value, _ := object.GetTagValueByKey(attr.Name)
            log.Debug("value", "view.NewCustomByObject: attribute =", attr.Name, ", value =", value)
            custom.SetValueByAttributeName(attr.Name, value)
        } else {
            for i := 0; i < vObject.NumField(); i++ {
                field := vObject.Type().Field(i)
                name := field.Tag.Get("json")

                if attr.Name == name {
                    value := vObject.FieldByName(field.Name)
                    if field.Type.Name() == "string" {
                        log.Debug("value", "view.NewCustomByObject: attribute =", attr.Name, ", value =", value.String())
                        custom.SetValueByAttributeName(attr.Name, value.String())
                    } else if field.Type.Name() == "bool" {
                        log.Debug("value", "view.NewCustomByObject: attribute =", attr.Name, ", value =", value.Bool())
                        custom.SetValueByAttributeName(attr.Name, value.Bool())
                    } else if field.Type.Name() == "int64" {
                        log.Debug("value", "view.NewCustomByObject: attribute =", attr.Name, ", value =", value.Int())
                        custom.SetValueByAttributeName(attr.Name, value.Int())
                    }
                    break
                }
            }
        }
    }

    log.Debug("success", "view.NewCustomByObject:", fmt.Sprintf("%+v", custom))
    return custom, nil
}

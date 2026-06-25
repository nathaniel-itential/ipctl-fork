// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package runners

import (
	"errors"
	"fmt"

	"github.com/itential/ipctl/internal/config"
	"github.com/itential/ipctl/internal/flags"
	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/internal/utils"
	"github.com/itential/ipctl/pkg/client"
	"github.com/itential/ipctl/pkg/resources"
	"github.com/itential/ipctl/pkg/services"
)

type GroupRunner struct {
	resource resources.GroupResourcer
	BaseRunner
}

func NewGroupRunner(c client.Client, cfg config.Provider) *GroupRunner {
	return &GroupRunner{
		resource:   resources.NewGroupResource(services.NewGroupService(c)),
		BaseRunner: NewBaseRunner(c, cfg),
	}
}

//////////////////////////////////////////////////////////////////////////////
// Reader Interface
//

func (r *GroupRunner) Get(in Request) (*Response, error) {
	logging.Trace()

	groups, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	return &Response{
		Keys:   []string{"name", "description"},
		Object: groups,
	}, nil

}

func (r *GroupRunner) Describe(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	groups, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	var grp *services.Group

	for _, ele := range groups {
		if ele.Name == name {
			grp = &ele
			break
		}
	}

	if grp == nil {
		return nil, errors.New(
			fmt.Sprintf("Group with name `%s` does not exist", name),
		)
	}

	return &Response{
		Object: grp,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Writer Interface
//

func (r *GroupRunner) Create(in Request) (*Response, error) {
	logging.Trace()

	var options flags.GroupCreateOptions
	utils.LoadObject(in.Options, &options)

	group := services.NewGroup(in.Args[0], options.Description)

	res, err := r.resource.Create(group)
	if err != nil {
		return nil, err
	}

	return &Response{
		Text:   fmt.Sprintf("Successfully created group `%s`", in.Args[0]),
		Object: res,
	}, nil
}

func (r *GroupRunner) Delete(in Request) (*Response, error) {
	logging.Trace()

	group, err := r.resource.GetByName(in.Args[0])
	if err != nil {
		return nil, err
	}

	if group.Provenance != "Pronghorn" {
		return nil, errors.New("cannot delete non-local group")
	}

	if err := r.resource.Delete(group.Id); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully deleted group `%s`", group.Name),
	}, nil
}

func (r *GroupRunner) Clear(in Request) (*Response, error) {
	logging.Trace()

	groups, err := r.resource.GetAll()
	if err != nil {
		return nil, err
	}

	var cnt int = 0

	for _, ele := range groups {
		if ele.Provenance == "Pronghorn" {
			if err := r.resource.Delete(ele.Id); err != nil {
				return nil, err
			}
			cnt++
		}
	}

	return &Response{
		Text: fmt.Sprintf("Successfully deleted %v group(s)", cnt),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Copier Interface
//

func (r *GroupRunner) Copy(in Request) (*Response, error) {
	logging.Trace()

	res, err := Copy(CopyRequest{Request: in, Type: "group"}, r)
	if err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully copied group `%s` from `%s` to `%s`", res.Name, res.From, res.To),
	}, nil
}

func (r *GroupRunner) CopyFrom(profile, name string) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	svc := services.NewGroupService(client)
	res := resources.NewGroupResource(svc)

	group, err := res.GetByName(name)
	if err != nil {
		return nil, err
	}

	return *group, err
}

func (r *GroupRunner) CopyTo(profile string, in any, replace bool) (any, error) {
	logging.Trace()

	client, cancel, err := NewClient(profile, r.config)
	if err != nil {
		return nil, err
	}
	defer cancel()

	svc := services.NewGroupService(client)

	name := in.(services.Group).Name

	if exists, err := svc.GetByName(name); exists != nil {
		if !replace {
			return nil, errors.New(fmt.Sprintf("group `%s` exists on the destination server, use --replace to overwrite", name))
		} else if err != nil {
			return nil, err
		}
		if err := svc.Delete(name); err != nil {
			return nil, err
		}
	}

	res, err := svc.Create(in.(services.Group))
	if err != nil {
		return nil, err
	}

	return res, nil

}

//////////////////////////////////////////////////////////////////////////////
// Importer Interface
//

func (r *GroupRunner) Import(in Request) (*Response, error) {
	logging.Trace()

	common := in.Common.(*flags.AssetImportCommon)

	var grp services.Group

	if err := importUnmarshalFromRequest(in, &grp); err != nil {
		return nil, err
	}

	if err := r.importGroup(grp, common.Replace); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully imported group `%s`", grp.Name),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Exporter Interface
//

func (r *GroupRunner) Export(in Request) (*Response, error) {
	logging.Trace()

	name := in.Args[0]

	grp, err := r.resource.GetByName(name)
	if err != nil {
		return nil, err
	}

	fn := fmt.Sprintf("%s.group.json", name)

	if err := exportAssetFromRequest(in, grp, fn); err != nil {
		return nil, err
	}

	return &Response{
		Text: fmt.Sprintf("Successfully exported gropu `%s` (%s)", grp.Name, grp.Id),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////
// Private functions
//

func (r *GroupRunner) importGroup(in services.Group, replace bool) error {
	logging.Trace()

	existing, err := r.resource.GetByName(in.Name)

	if err != nil {
		if !errors.Is(err, resources.ErrNotFound) {
			return err
		}
	}

	if existing != nil {
		if replace {
			if err := r.resource.Delete(existing.Id); err != nil {
				return err
			}
		} else {
			return errors.New(
				fmt.Sprintf("group `%s` already exists, use --replace to overwrite it", in.Name),
			)
		}
	}

	_, err = r.resource.Create(in)

	return err

}

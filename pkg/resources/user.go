package resources

import (
	"context"
	"strings"

	"github.com/gthesheep/terraform-provider-tableau/pkg/tableau"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	userSiteRoles = []string{
		"Creator",
		"Explorer",
		"ExplorerCanPublish",
		"SiteAdministratorExplorer",
		"SiteAdministratorCreator",
		"Unlicensed",
		"Viewer",
	}

	userAuthSettings = []string{
		"SAML",
		"ServerDefault",
		"OpenID",
		"TABID_WITH_MFA",
	}
)

var userSchema = map[string]*schema.Schema{
	"email": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "User email",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "User name",
	},
	"full_name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Display name for user",
	},
	"site_role": &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Site role for the user",
		ValidateFunc: validation.StringInSlice(userSiteRoles, false),
	},
	"auth_setting": &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Auth setting for the user",
		ValidateFunc: validation.StringInSlice(userAuthSettings, false),
	},
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Schema: userSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tableau.Client)

	var diags diag.Diagnostics

	userID := d.Id()

	user, err := c.GetUser(userID)
	if err != nil {
		// TODO: this is terrible
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", user.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", user.FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("site_role", user.SiteRole); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_setting", user.AuthSetting); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tableau.Client)

	var diags diag.Diagnostics

	email := d.Get("email").(string)
	name := d.Get("name").(string)
	fullName := d.Get("full_name").(string)
	siteRole := d.Get("site_role").(string)
	authSetting := d.Get("auth_setting").(string)

	u, err := c.CreateUser(email, name, fullName, siteRole, authSetting)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*u.ID)

	_, err = c.UpdateUser(*u.ID, fullName, email, siteRole, authSetting)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tableau.Client)
	userId := d.Id()

	if d.HasChange("full_name") || d.HasChange("email") || d.HasChange("site_role") || d.HasChange("auth_setting") {
		fullName := d.Get("full_name").(string)
		email := d.Get("email").(string)
		siteRole := d.Get("site_role").(string)
		authSetting := d.Get("auth_setting").(string)

		_, err := c.UpdateUser(userId, fullName, email, siteRole, authSetting)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tableau.Client)
	userId := d.Id()

	var diags diag.Diagnostics

	_, err := c.DeleteUser(userId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

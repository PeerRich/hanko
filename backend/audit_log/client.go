package auditlog

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type Client interface {
	Create(echo.Context, models.AuditLogType, *models.User, error) error
}

type client struct {
	persister     persistence.Persister
	enableStoring bool
}

func NewClient(persister persistence.Persister, config config.AuditLog) Client {
	return &client{
		persister:     persister,
		enableStoring: config.EnableStoring,
	}
}

func (c *client) Create(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	var err error = nil
	if c.enableStoring {
		err = c.store(context, auditLogType, user, logError)
		if err != nil {
			return err
		}
	}

	// TODO: log each auditLogType (logrus, zerolog, ...)

	return nil
}

func (c *client) store(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create id: %w", err)
	}
	var userId *uuid.UUID = nil
	var userEmail = ""
	if user != nil {
		userId = &user.ID
		userEmail = user.Email
	}
	errString := ""
	if logError != nil {
		// check if error is not nil, because else the string (formatted with fmt.Sprintf) would not be empty but look like this: `%!s(<nil>)`
		errString = fmt.Sprintf("%s", logError)
	}
	e := models.AuditLog{
		ID:                id,
		Type:              auditLogType,
		Error:             errString,
		MetaHttpRequestId: context.Response().Header().Get(echo.HeaderXRequestID),
		MetaUserAgent:     context.Request().UserAgent(),
		MetaSourceIp:      context.RealIP(),
		ActorUserId:       userId,
		ActorEmail:        userEmail,
	}

	return c.persister.GetAuditLogPersister().Create(e)
}

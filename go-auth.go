package main

import "C"

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"plugin"

	bes "./backends"
)

type Backend interface {
	GetUser(username, password string) bool
	GetSuperuser(username string) bool
	CheckAcl(username, topic, clientId string, acc int32) bool
	GetName() string
	Halt()
}

type CommonData struct {
	Backends         map[string]Backend
	Plugin           *plugin.Plugin
	PInit            func(map[string]string, log.Level) error
	PGetName         func() string
	PGetUser         func(username, password string) bool
	PGetSuperuser    func(username string) bool
	PCheckAcl        func(username, topic, clientid string, acc int) bool
	PHalt            func()
	Superusers       []string
	CheckPrefix      bool
	Prefixes         map[string]string
	LogLevel         log.Level
}

var allowedBackends = map[string]bool{
	"postgres": true,
	"plugin":   true,
}

var backends []string
var authOpts map[string]string
var commonData CommonData

//export AuthPluginInit
func AuthPluginInit(keys []string, values []string, authOptsNum int) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	superusers := make([]string, 10, 10)

	cmbackends := make(map[string]Backend)

	//Initialize common struct with default and given values
	commonData = CommonData{
		Superusers:       superusers,
		CheckPrefix:      false,
		Prefixes:         make(map[string]string),
		LogLevel:         log.InfoLevel,
	}

	//First, get backends
	backendsOk := false
	authOpts = make(map[string]string)
	for i := 0; i < authOptsNum; i++ {
		if keys[i] == "backends" {
			backends = strings.Split(strings.Replace(values[i], " ", "", -1), ",")
			if len(backends) > 0 {
				backendsCheck := true
				for _, backend := range backends {
					if _, ok := allowedBackends[backend]; !ok {
						backendsCheck = false
						log.Errorf("backend not allowed: %s", backend)
					}
				}
				backendsOk = backendsCheck
			}
		} else {
			authOpts[keys[i]] = values[i]
		}
	}

	//Log and end program if backends are wrong
	if !backendsOk {
		log.Fatal("\nbackends error\n")
	}

	//Check if log level is given. Set level if any valid option is given.
	if logLevel, ok := authOpts["log_level"]; ok {

		logLevel = strings.Replace(logLevel, " ", "", -1)

		if logLevel == "debug" {
			commonData.LogLevel = log.DebugLevel
		} else if logLevel == "info" {
			commonData.LogLevel = log.InfoLevel
		} else if logLevel == "warn" {
			commonData.LogLevel = log.WarnLevel
		} else if logLevel == "error" {
			commonData.LogLevel = log.ErrorLevel
		} else if logLevel == "fatal" {
			commonData.LogLevel = log.FatalLevel
		} else if logLevel == "panic" {
			commonData.LogLevel = log.PanicLevel
		}

	}

	//Initialize backends
	for _, bename := range backends {
		var beIface Backend
		var bErr error

		if bename == "plugin" {
			plug, plErr := plugin.Open(authOpts["plugin_path"])
			if plErr != nil {
				log.Errorf("Could not init custom plugin: %s", plErr)
				commonData.Plugin = nil
			} else {
				commonData.Plugin = plug

				plInit, piErr := commonData.Plugin.Lookup("Init")

				if piErr != nil {
					log.Errorf("Couldn't find func Init in plugin: %s", plErr)
					commonData.Plugin = nil
					continue
				}

				initFunc := plInit.(func(authOpts map[string]string, logLevel log.Level) error)

				ipErr := initFunc(authOpts, commonData.LogLevel)
				if ipErr != nil {
					log.Errorf("Couldn't init plugin: %s", ipErr)
					commonData.Plugin = nil
					continue
				}

				commonData.PInit = initFunc

				plName, gErr := commonData.Plugin.Lookup("GetName")

				if gErr != nil {
					log.Errorf("Couldn't find func GetName in plugin: %s", gErr)
					commonData.Plugin = nil
					continue
				}

				nameFunc := plName.(func() string)
				commonData.PGetName = nameFunc

				plGetUser, pgErr := commonData.Plugin.Lookup("GetUser")

				if pgErr != nil {
					log.Errorf("Couldn't find func GetUser in plugin: %s", pgErr)
					commonData.Plugin = nil
					continue
				}

				getUserFunc := plGetUser.(func(username, password string) bool)
				commonData.PGetUser = getUserFunc

				if pgErr != nil {
					log.Errorf("Couldn't find func GetUser in plugin: %s", pgErr)
					commonData.Plugin = nil
					continue
				}

				plGetSuperuser, psErr := commonData.Plugin.Lookup("GetSuperuser")

				if psErr != nil {
					log.Errorf("Couldn't find func GetSuperuser in plugin: %s", psErr)
					commonData.Plugin = nil
					continue
				}

				getSuperuserFunc := plGetSuperuser.(func(username string) bool)
				commonData.PGetSuperuser = getSuperuserFunc

				plCheckAcl, pcErr := commonData.Plugin.Lookup("CheckAcl")

				if pcErr != nil {
					log.Errorf("Couldn't find func CheckAcl in plugin: %s", pcErr)
					commonData.Plugin = nil
					continue
				}

				checkAclFunc := plCheckAcl.(func(username, topic, clientid string, acc int) bool)
				commonData.PCheckAcl = checkAclFunc

				plHalt, phErr := commonData.Plugin.Lookup("Halt")

				if phErr != nil {
					log.Errorf("Couldn't find func Halt in plugin: %s", phErr)
					commonData.Plugin = nil
					continue
				}

				haltFunc := plHalt.(func())
				commonData.PHalt = haltFunc

				log.Infof("Backend registered: %s", commonData.PGetName())

			}
		} else {
			if bename == "postgres" {
				beIface, bErr = bes.NewPostgres(authOpts, commonData.LogLevel)
				if bErr != nil {
					log.Fatalf("Backend register error: couldn't initialize %s backend with error %s.", bename, bErr)
				} else {
					log.Infof("Backend registered: %s", beIface.GetName())
					cmbackends["postgres"] = beIface.(bes.Postgres)
				}
			}
		}

	}

	if checkPrefix, ok := authOpts["check_prefix"]; ok && strings.Replace(checkPrefix, " ", "", -1) == "true" {
		//Check that backends match prefixes.
		if prefixesStr, ok := authOpts["prefixes"]; ok {
			prefixes := strings.Split(strings.Replace(prefixesStr, " ", "", -1), ",")
			if len(prefixes) == len(backends) {
				//Set prefixes
				for i, backend := range backends {
					commonData.Prefixes[prefixes[i]] = backend
				}
				log.Infof("Prefixes enabled for backends %s with prefixes %s.", authOpts["backends"], authOpts["prefixes"])
				commonData.CheckPrefix = true
			} else {
				log.Errorf("Error: got %d backends and %d prefixes, defaulting to prefixes disabled.", len(backends), len(prefixes))
				commonData.CheckPrefix = false
			}

		} else {
			log.Warn("Error: prefixes enabled but no options given, defaulting to prefixes disabled.")
			commonData.CheckPrefix = false
		}
	} else {
		commonData.CheckPrefix = false
	}

	commonData.Backends = cmbackends

}

//export AuthUnpwdCheck
func AuthUnpwdCheck(username, password string) bool {

	//Loop through backends checking for user.

	authenticated := false

	//If prefixes are enabled, checkt if username has a valid prefix and use the correct backend if so.
	if commonData.CheckPrefix {
		validPrefix, bename := CheckPrefix(username)
		if validPrefix {

			if bename == "plugin" {
				authenticated = CheckPluginAuth(username, password)
			} else {

				var backend = commonData.Backends[bename]

				if backend.GetUser(username, password) {
					authenticated = true
					log.Debugf("user %s authenticated with backend %s", username, backend.GetName())
				}

			}

		} else {
			//If there's no valid prefix, check all backends.
			authenticated = CheckBackendsAuth(username, password)
			//If not authenticated, check for a present plugin
			if !authenticated {
				authenticated = CheckPluginAuth(username, password)
			}
		}
	} else {
		authenticated = CheckBackendsAuth(username, password)
		//If not authenticated, check for a present plugin
		if !authenticated {
			authenticated = CheckPluginAuth(username, password)
		}
	}

	return authenticated
}

//export AuthAclCheck
func AuthAclCheck(clientid, username, topic string, acc int) bool {

	aclCheck := false

	//If prefixes are enabled, checkt if username has a valid prefix and use the correct backend if so.
	//Else, check all backends.
	if commonData.CheckPrefix {
		validPrefix, bename := CheckPrefix(username)
		if validPrefix {

			if bename == "plugin" {

				aclCheck = CheckPluginAcl(username, topic, clientid, acc)

			} else {

				var backend = commonData.Backends[bename]

				log.Debugf("Superuser check with backend %s", backend.GetName())
				if backend.GetSuperuser(username) {
					log.Debugf("superuser %s acl authenticated with backend %s", username, backend.GetName())
					aclCheck = true
				}

				//If not superuser, check acl.
				if !aclCheck {
					log.Debugf("Acl check with backend %s", backend.GetName())
					if backend.CheckAcl(username, topic, clientid, int32(acc)) {
						log.Debugf("user %s acl authenticated with backend %s", username, backend.GetName())
						aclCheck = true
					}
				}
			}

		} else {
			//If there's no valid prefix, check all backends.
			aclCheck = CheckBackendsAcl(username, topic, clientid, acc)
			//If acl hasn't passed, check for plugin.
			if !aclCheck {
				aclCheck = CheckPluginAcl(username, topic, clientid, acc)
			}
		}
	} else {
		aclCheck = CheckBackendsAcl(username, topic, clientid, acc)
		//If acl hasn't passed, check for plugin.
		if !aclCheck {
			aclCheck = CheckPluginAcl(username, topic, clientid, acc)
		}
	}

	log.Debugf("Acl is %s for user %s", aclCheck, username)

	return aclCheck
}

//export AuthPskKeyGet
func AuthPskKeyGet() bool {
	return true
}

//CheckPrefix checks if a username contains a valid prefix. If so, returns ok and the suitable backend name; else, !ok and empty string.
func CheckPrefix(username string) (bool, string) {
	if strings.Index(username, "_") > 0 {
		userPrefix := username[0:strings.Index(username, "_")]
		if prefix, ok := commonData.Prefixes[userPrefix]; ok {
			log.Debugf("Found prefix for user %s, using backend %s.", username, prefix)
			return true, prefix
		}
	}
	return false, ""
}

//CheckBackendsAuth checks for all backends if a username is authenticated and sets the authenticated param.
func CheckBackendsAuth(username, password string) bool {

	authenticated := false

	for _, bename := range backends {

		if bename == "plugin" {
			continue
		}

		var backend = commonData.Backends[bename]

		log.Debugf("checking user %s with backend %s", username, backend.GetName())

		if backend.GetUser(username, password) {
			authenticated = true
			log.Debugf("user %s authenticated with backend %s", username, backend.GetName())
			break
		}
	}

	return authenticated

}

//CheckBackendsAcl  checks for all backends if a username is superuser or has acl rights and sets the aclCheck param.
func CheckBackendsAcl(username, topic, clientid string, acc int) bool {
	//Check superusers first

	aclCheck := false

	for _, bename := range backends {

		if bename == "plugin" {
			continue
		}

		var backend = commonData.Backends[bename]

		log.Debugf("Superuser check with backend %s", backend.GetName())
		if backend.GetSuperuser(username) {
			log.Debugf("superuser %s acl authenticated with backend %s", username, backend.GetName())
			aclCheck = true
			break
		}
	}

	if !aclCheck {
		for _, bename := range backends {

			if bename == "plugin" {
				continue
			}

			var backend = commonData.Backends[bename]

			log.Debugf("Acl check with backend %s", backend.GetName())
			if backend.CheckAcl(username, topic, clientid, int32(acc)) {
				log.Debugf("user %s acl authenticated with backend %s", username, backend.GetName())
				aclCheck = true
				break
			}
		}
	}

	return aclCheck

}

//CheckPluginAuth checks that the plugin is not nil and returns the plugins auth response.
func CheckPluginAuth(username, password string) bool {
	if commonData.Plugin != nil {
		return commonData.PGetUser(username, password)
	}
	return false
}

//CheckPluginAcl checks that the plugin is not nil and returns the superuser/acl response.
func CheckPluginAcl(username, topic, clientid string, acc int) bool {
	if commonData.Plugin != nil {
		aclCheck := commonData.PGetSuperuser(username)
		if !aclCheck {
			aclCheck = commonData.PCheckAcl(username, topic, clientid, acc)
		}
	}
	return false
}

//export AuthPluginCleanup
func AuthPluginCleanup() {
	log.Info("Cleaning up plugin")

	//Halt every registered backend.

	for _, v := range commonData.Backends {
		v.Halt()
	}

	if commonData.Plugin != nil {
		commonData.PHalt()
	}
}

func main() {}

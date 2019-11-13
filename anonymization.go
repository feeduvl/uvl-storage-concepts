package main

import (
	"encoding/json"
	"regexp"
)

type Anonymization struct {
	tag     string
	pattern *regexp.Regexp
}

var anonymizations = []Anonymization{
	{"NUMBER", regexp.MustCompile("(?:(?:#|\\bn)\\d{10}|\\b(?:\\d{3,4}|\\d{7,12})\\b)")},
	{"FISCALCODE", regexp.MustCompile("(?i)\\bn?[a-z]{6}[0-9]{2}[a-z][0-9]{2}[a-z][0-9]{3}[a-z]\\b")},
	{"PERSON", regexp.MustCompile("(?i)\\b(?:leo|ugo|ada|eva|isa|ida|lia|zoe|aldo|ciro|dino|elia|elio|enzo|ezio|gino|gigi|igor|ivan|lapo|luca|nino|omar|vito|alba|anna|asia|dina|dora|elsa|emma|gaia|irma|ines|lisa|lara|mina|mara|olga|rosa|rita|sara|vera|tina|boris|bruno|carlo|dario|diego|fabio|guido|ivano|livio|loris|lucio|luigi|marco|mario|mauro|mirko|oscar|paolo|piero|renzo|rocco|romeo|salvo|santo|adele|agata|alice|ambra|anita|bruna|carla|clara|daria|delia|diana|elena|elisa|ester|flora|greta|gilda|giada|gemma|ivana|irene|luisa|lucia|luana|livia|linda|lidia|licia|laura|katia|moira|mimma|maura|marta|maria|magda|nadia|noemi|norma|piera|paola|sofia|sonia|viola|adolfo|amedeo|angelo|andrea|arturo|biagio|cesare|danilo|davide|donato|egidio|emilio|enrico|ettore|fausto|flavio|fulvio|gianni|giulio|iacopo|ilario|marino|matteo|mattia|moreno|nicola|nicole|nunzio|orazio|oreste|pietro|renato|romano|sandro|savino|sergio|silvio|simone|ubaldo|valter|walter|agnese|amalia|angela|aurora|bianca|carmen|chiara|cinzia|clelia|debora|donata|elvira|emilia|enrica|fulvia|franca|flavia|fatima|fausta|grazia|gloria|giulia|ilaria|ilenia|ileana|luigia|lorena|monica|miriam|milena|marisa|marina|marica|oriana|pamela|romina|romana|renata|sandra|serena|silvia|simona|stella|teresa|tamara|adriano|alberto|alessio|alfredo|alfonso|antonio|armando|arnaldo|attilio|augusto|aurelio|camillo|carmelo|carmine|claudio|corrado|damiano|daniele|edoardo|ermanno|ernesto|eugenio|filippo|gaetano|gennaro|gerardo|germano|giacomo|giorgio|ignazio|lorenzo|luciano|manuele|mariano|martino|michele|orlando|osvaldo|ottavio|roberto|rodolfo|rosario|ruggero|samuele|saverio|silvano|stefano|tiziano|tommaso|umberto|valerio|adriana|alessia|antonia|arianna|assunta|barbara|camilla|carmela|cecilia|claudia|celeste|daniela|deborah|diletta|erminia|ernesta|evelina|gaspare|fabiana|fabiola|gisella|giorgia|ginevra|luciana|lorenza|lorella|liliana|letizia|lavinia|mirella|miranda|michela|melissa|matilde|martina|manuela|mafalda|ornella|rossana|rosanna|roberta|rebecca|rachele|sabrina|samanta|silvana|susanna|viviana|vanessa|valeria|tiziana|agostino|ambrogio|battista|bernardo|calogero|clemente|cristian|costanzo|domenico|emanuele|emiliano|fabrizio|federico|fernando|gabriele|giacinto|gianluca|gilberto|giordano|giovanni|girolamo|giuliano|giuseppe|gregorio|leonardo|lodovico|ludovico|marcello|maurizio|pasquale|patrizio|raffaele|raimondo|riccardo|serafino|tarcisio|vincenzo|virgilio|virginio|vittorio|angelica|beatrice|carolina|caterina|concetta|cristina|costanza|carlotta|domenica|eleonora|emanuela|emiliana|fabrizia|floriana|fiorella|fernanda|federica|giuliana|giuditta|giovanna|isabella|ludovica|lucrezia|loredana|lodovica|marilena|mariella|marianna|marcella|patrizia|rossella|samantha|serafina|stefania|vittoria|virginia|vincenza|veronica|antonello|beniamino|benedetto|cristiano|christian|fortunato|francesco|giancarlo|gianluigi|gianpiero|gianpaolo|gianmarco|gianmaria|guglielmo|lanfranco|pierluigi|salvatore|valentino|annamaria|antonella|benedetta|cristiana|donatella|francesca|graziella|gabriella|marinella|maddalena|nicoletta|raffaella|simonetta|valentina|alessandro|bartolomeo|ferdinando|gianfranco|gianpietro|gioacchino|pierangelo|sebastiano|alessandra|annunziata|antonietta|elisabetta|giuseppina|margherita|piergiorgio|giambattista|giandomenico|gianbattista|massimiliano|michelangelo)\\b")},
}

func AnonymizeString(s string) string {
	result := s
	for _, anonymization := range anonymizations {
		result = anonymization.pattern.ReplaceAllLiteralString(result, "["+anonymization.tag+"]")
	}
	return result
}

type AnonymizedString string

func (a AnonymizedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(AnonymizeString(string(a)))
}

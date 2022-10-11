package gtmcdc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
)

// Config stores the configurations for the filter
type Config struct {
	KafkaBrokerList string `env:"GTMCDC_KAFKA_BROKERS" envDefault:"off"`
	KafkaTopic      string `env:"GTMCDC_KAFKA_TOPIC" envDefault:"cdc-test"`
	PromHTTPAddr    string `env:"GTMCDC_PROM_HTTP_ADDR" envDefault:"off"`
	LogFile         string `env:"GTMCDC_LOG" envDefault:"stderr"`
	LogLevel        string `env:"GTMCDC_LOG_LEVEL" envDefault:"debug"`
}

// LoadConfig loads the filter configurations from file
func LoadConfig(envFile string) *Config {
	// environment varialbe overrides command
	env2 := os.Getenv("GTMCDC_ENV")
	if envFile == "" && env2 != "" {
		envFile = env2
	}

	if envFile != "" {
		_ = godotenv.Load(envFile)
	}

	conf := Config{}
	if err := env.Parse(&conf); err != nil {
		log.Warnf("%+v", err)
	}

	return &conf
}

// DoFilter is the main processing loop that
// reads journal extract and publish messages
func DoFilter(fin, fout *os.File, producer *Producer, metrics *Metrics) {
	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := scanner.Text()
		metrics.IncrCounter("lines_read_from_input")

		// log with fields
		logf := log.WithFields(log.Fields{"journal": line, "func": "Filter.DoFilter"})

		rec, err := Parse(line)
		if err != nil {
			logf.Errorf("Unable to parse record")
			metrics.IncrCounter("lines_parse_error")
			continue
		} else {
			metrics.IncrCounter("lines_parsed")

			jsonstr, err := rec.JSON()
			if err != nil {
				logf.Fatalf("cannot marshal to JSON due to %+v", err)
			}

			logf.Debugf("line parsed to json %s", jsonstr)

			if producer.IsKafkaAvailable() && jsonstr != "" {
				start := time.Now()

				err = producer.PublishMessage(jsonstr)
				if err != nil {
					logf.Warnf("Unable to publish message for journal record. %+v", err)
					metrics.IncrCounter("lines_parsed_but_not_published")
				} else {
					metrics.IncrCounter("lines_parsed_and_published")
					elapsed := time.Since(start)
					metrics.HistoObserve("message_publish_to_kafka", float64(elapsed/time.Microsecond))
				}
			}

			// send to output only after a message is successfully published
			_, err = fmt.Fprintln(fout, line)
			if err != nil {
				metrics.IncrCounter("lines_output_write_error")
				logf.Errorf("Unable to write to output")
			} else {
				metrics.IncrCounter("lines_output_written")
			}
		}
	}
}

// InitLogging initialize log output based on configuration
func InitLogging(logFile, logLevel string) {
	var file *os.File
	var err error

	// force log to file if log is set to stderr and stdout is a tty
	// so that the log messages doesn't mix with stdout and make a mess
	// on the screen
	if logFile == "stderr" && isatty.IsTerminal(os.Stdout.Fd()) {
		logFile = "cdcfilter.log"
	}

	if strings.EqualFold(logFile, "stderr") {
		file = os.Stderr
		log.SetFormatter(&log.TextFormatter{})
	} else {
		file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("can't open log file for write.")
		}
		log.SetFormatter(&log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyTime: "@timestamp",
				log.FieldKeyMsg:  "@message",
			}})
	}

	log.SetOutput(file)

	level, err := log.ParseLevel(logLevel)
	if err == nil {
		log.SetLevel(level)
	}
}

// InitInputAndOutput initialize input and output files for the filter
func InitInputAndOutput(inputFile, outputFile string) (*os.File, *os.File) {
	// initialize input and output files for this filter
	fin := os.Stdin
	fout := os.Stdout
	var err error

	// the primary reason for this read input from file logic is that
	// in my Goland IDE there's no way to redirect STDIN and STDOUT
	// in Run configuration, so I couldn't debug the code without this.
	if inputFile != "" && !strings.EqualFold(inputFile, "stdin") {
		fin, err = os.OpenFile(inputFile, os.O_RDONLY, 0666)
		if err != nil {
			log.Fatalf("Unable to input file %s", inputFile)
		} else {
			log.Debugf("Input file: %s", inputFile)
		}
	}

	if outputFile != "" && !strings.EqualFold(outputFile, "stdout") {
		fout, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to output file %s", outputFile)
		} else {
			log.Debugf("Output file: %s", outputFile)
		}
	}

	return fin, fout
}

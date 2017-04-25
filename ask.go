package input

import (
	"bytes"
	"fmt"
)

// Ask asks the user for input using the given query. The response is
// returned as string. Error is returned based on the given option.
// If Loop is true, it continue to ask until it receives valid input.
//
// If the user sends SIGINT (Ctrl+C) while reading input, it catches
// it and return it as a error.
func (i *UI) Ask(query string, opts *Options) (string, error) {
	i.once.Do(i.setDefault)

	// Display the query to the user.
	fmt.Fprintf(i.Writer, "%s%s", opts.Prefix, query)

	// resultStr and resultErr are return val of this function
	var resultStr string
	var resultErr error

	var defaultErrSuffix = "\n\n"

	loopCount := 0
	for {
		loopCount++

		// Construct the instruction to user.
		var buf bytes.Buffer
		if !opts.HideOrder {
			buf.WriteString(fmt.Sprintf("\n%sEnter a value", opts.Prefix))
		}

		if loopCount > 1 {
			defaultLoopOrder := fmt.Sprintf("\n%sEnter a value", opts.Prefix)
			if opts.LoopOrder != defaultLoopOrder && opts.LoopOrder != "" {
				defaultLoopOrder = fmt.Sprintf("\n%s%s", opts.Prefix, opts.LoopOrder)
			}
			buf.WriteString(defaultLoopOrder)
		}

		if opts.Default != "" && !opts.HideDefault {
			defaultVal := opts.Default
			if opts.MaskDefault {
				defaultVal = maskString(defaultVal)
			}
			buf.WriteString(fmt.Sprintf(" (Default is %s)", defaultVal))
		}

		// Display the instruction to user and ask to input.
		buf.WriteString(": ")
		fmt.Fprintf(i.Writer, buf.String())

		// Read user input from UI.Reader.
		line, err := i.read(opts.readOpts())
		if err != nil {
			resultErr = err
			break
		}

		if opts.ErrSuffix != "" && opts.ErrSuffix != defaultErrSuffix {
			defaultErrSuffix = opts.ErrSuffix
		}

		// line is empty but default is provided returns it
		if line == "" && opts.Default != "" {
			resultStr = opts.Default
			break
		}

		if line == "" && opts.Required {
			if !opts.Loop {
				resultErr = ErrEmpty
				break
			}

			fmt.Fprintf(i.Writer, "Input must not be empty.%s", defaultErrSuffix)
			continue
		}

		// validate input by custom fuction
		validate := opts.validateFunc()
		if err := validate(line); err != nil {
			if !opts.Loop {
				resultErr = err
				break
			}

			validateFuncErr := "Failed to validate input string: "

			if opts.HideValidateFuncErr {
				validateFuncErr = ""
			}

			fmt.Fprintf(i.Writer, "%s%s%s%s", opts.Prefix, validateFuncErr, err, defaultErrSuffix)
			continue
		}

		// Reach here means it gets ideal input.
		resultStr = line
		break
	}

	// Insert the new line for next output
	fmt.Fprintf(i.Writer, "\n")

	return resultStr, resultErr
}

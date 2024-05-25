package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"unicode/utf8"
)

var cmd_args []string
var input_file string
var has_input_arg bool = false
var output_name string = ""
var output_location string
var has_output_location bool = false
var image_alt string
var output_formats []string
var output_width string
var output_height string
var has_output_width bool = false
var has_output_height bool = false
var use_width_for_output string
var has_newline_for_properties bool = false
var absolute_path bool = false
var has_custom_path bool = false
var custom_path string
var indenting_amount int = 2
var has_mime bool = false
var avif_crf = "25"
var webp_quality = "83"
var expanded_args []string
var multiargument = regexp.MustCompile(`^-[a-zA-Z]{2,}$`)

func main(){
  for _, arg := range os.Args {
  	if multiargument.MatchString(arg){
      letters := arg[1:]
      for _, letter := range letters {
        expanded_args = append(expanded_args , string(letter))
      }
    } else {
      expanded_args = append(expanded_args, arg)
    }
  }

  for n, args := range expanded_args {
    switch {
      case args == "-i" || args == "--input":
        if n + 1 < len(expanded_args) {
          if _, err := os.Stat(expanded_args[n+1]); err == nil {
            input_file = expanded_args[n+1]
            has_input_arg = true
          } else {
            fmt.Println("The file provided doesn't exist");
            os.Exit(1)
          }
        } else {
          fmt.Println("You didn't provide an input file")
          os.Exit(1)
        }
      case args == "-o" || args == "--output":
        if n + 1 < len(expanded_args) {
          has_output_location = true
          output_location = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide an output location")
          os.Exit(1)
        }
      case args == "--width":
        if n + 1 < len(expanded_args) {
          has_output_width = true
          output_width = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a width")
          os.Exit(1)
        }
      case args == "-h" || args == "--height":
        if n + 1 < len(expanded_args) {
          has_output_height = true
          output_height = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a height")
          os.Exit(1)
        }
      case args == "-s" || args == "--schema":
        if n + 1 < len(expanded_args) {
          output_name = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a schema")
          os.Exit(1)
        }
      case args == "--ind" || args == "--indenting":
        if n + 1 < len(expanded_args) {
          indenting_amount, _  = strconv.Atoi(expanded_args[n+1])
        } else {
          fmt.Println("You didn't provide an indenting amount")
          os.Exit(1)
        }
      case args == "-p" || args == "--path":
        if n + 1 < len(expanded_args) {
          has_custom_path = true
          custom_path = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a path")
          os.Exit(1)
        }
      case args == "--alt":
        if n + 1 < len(expanded_args) {
          image_alt = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide an alt description")
          os.Exit(1)
        }
      case args == "-q" || args == "--quality":
        if n + 1 < len(expanded_args) {
          webp_quality = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a quality value for your webp image")
          os.Exit(1)
        }
      case args == "-c" || args == "--crf":
        if n + 1 < len(expanded_args) {
          avif_crf = expanded_args[n+1]
        } else {
          fmt.Println("You didn't provide a crf value for your avif image")
          os.Exit(1)
        }
      case args == "--mime" || args == "-m" || args == "m":
        has_mime = true
      case args == "--absolute" || args == "-t" || args == "t":
        absolute_path = true
      case args == "--newlines" || args == "-n" || args == "n":
        has_newline_for_properties = true
      case args == "--avif" || args == "-a" || args == "a":
        output_formats = append(output_formats, "avif")
      case args == "--webp" || args == "-w" || args == "w":
        output_formats = append(output_formats, "webp")
      case args == "--jpg" || args == "--jpeg" || args == "-j" || args == "j":
        output_formats = append(output_formats, "jpg")
    }
  }

  if has_input_arg {
    if has_output_location {
      output_code := construct_output_string()
      fmt.Println(output_code)
      run_command()
    } else {
      fmt.Println("You need to provide an output location")
    }
  } else {
    fmt.Println("You need to provide an input file")
  }
}

func indent_code(indenting int) string{
  spaces := ""
  for i := 0; i < indenting; i++{
    spaces += " "
  }
  return spaces
}

func regex_src(str string, to_include string) string{
  src_regex := regexp.MustCompile(`src="`)
  srcset_regex := regexp.MustCompile(`srcset="`)

  src_replace := `src="`
  srcset_replace := `srcset="`

  str = src_regex.ReplaceAllString(str, src_replace + to_include)
  str = srcset_regex.ReplaceAllString(str, srcset_replace + to_include)

  return str
}
func trimLastChar(s string) string {
  r, size := utf8.DecodeLastRuneInString(s)
  if r == utf8.RuneError && (size == 0 || size == 1) {
    size = 0
  }
  return s[:len(s)-size]
}

func initial_formatting(str string, indent_amount int) string{
  new_line_regex := regexp.MustCompile(`>`)
  str = new_line_regex.ReplaceAllString(str, ">\n")

  indenting_source_regex := regexp.MustCompile(`<s`)
  indenting_img_regex := regexp.MustCompile(`<i`)

  str = indenting_source_regex.ReplaceAllString(str, indent_code(indent_amount) + `<s`)
  str = indenting_img_regex.ReplaceAllString(str, indent_code(indent_amount) + `<i`)

  str = trimLastChar(str)

  return str
}

func delete_mime(str string) string{
  mime_regex := regexp.MustCompile(` type.*2?"`)
  str = mime_regex.ReplaceAllString(str, "") 
  return str
}

func add_custom_path(str string, custom_path string) string{
  str = regex_src(str, custom_path + "/")
  return str
}

func add_absolute_paths(str string) string{
  str = regex_src(str, "/")
  return str
}

func newline_for_properties(str string, indent_amount int) string{
  newline_regex := regexp.MustCompile(` `)
  str = newline_regex.ReplaceAllString(str, "\n" + indent_code(indent_amount * 2))
  return str
}

func construct_output_string() string{
  var output_str string
  var path string

  if output_name != "" {
    path = output_name
  } else {
    path = output_location
  }
  switch len(output_formats){
    case 1:
     break
    case 2,3:
      sort.Strings(output_formats)
      for i := 0; i < len(output_formats); i++{
        if output_formats[i] == "jpg"{
          output_formats = append(output_formats[:i],output_formats[i+1:]...)
          output_formats = append(output_formats, "jpg")
        }
      }
  }

  switch(len(output_formats)){
    case 0:
      return "Please provide atleast one output format"
    case 1:
      output_str = `<img src="` + path + "." + output_formats[0] + `" alt="` + image_alt + `">`
      break;
    case 2:
      output_str = `<picture><source srcset="`+ path + "." + output_formats[0] + `" type="image/` + output_formats[0]  +  `"><img src="` + path + "." + output_formats[1]  + `" alt="` + image_alt + `"></picture>`
  if has_newline_for_properties {
    output_str = newline_for_properties(output_str, indenting_amount)
  }
      output_str = initial_formatting(output_str, indenting_amount)
      break
    case 3:
      output_str = `<picture><source srcset="`+ path + "." + output_formats[0] + `" type="image/` + output_formats[0]  +  `"><source srcset="`+ path + "." + output_formats[1] + `" type="image/` + output_formats[1]  +  `"><img src="` + path + "." + output_formats[2]  + `" alt="` + image_alt + `"></picture>`
  if has_newline_for_properties {
    output_str = newline_for_properties(output_str, indenting_amount)
  }
      output_str = initial_formatting(output_str, indenting_amount)
      break
  }
  

  if has_custom_path {
    output_str = add_custom_path(output_str, custom_path)
  }

  if absolute_path {
    output_str = add_absolute_paths(output_str)
  }

  if !has_mime {
    output_str = delete_mime(output_str)
  }

  return output_str
}

func run_command(){

  args := []string {"-i", input_file}

  if has_output_height && has_output_width {
    args = append(args, "-vf", "scale=" + output_width + ":" + output_height)
  } else if has_output_width {
    args = append(args, "-vf", "scale=" + output_width + ":-1")
  } else if has_output_height {
    args = append(args, "-vf", "scale=" +  "-1:" + output_height)
  }

  for i := 0; i < len(output_formats); i++ {
    loop_args := args

    switch(output_formats[i]){
      case "avif":
        loop_args = append(loop_args, "-still-picture", "1", "-pix_fmt", "yuv420p", "-crf", avif_crf)
      case "webp":
        loop_args = append(loop_args, "-quality", webp_quality)
    }

    loop_args = append(loop_args, output_location + "." + output_formats[i])

    ffmpeg_command := exec.Command("ffmpeg", loop_args...)
    _, err := ffmpeg_command.Output()

    if err != nil {
      fmt.Println(err)
    }
  }
}

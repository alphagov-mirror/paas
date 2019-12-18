SPEC_DIR = File.expand_path(__dir__)
CONCOURSE_DIR = File.expand_path(File.join(SPEC_DIR, '..'))
TASKS_DIR = File.join(CONCOURSE_DIR, 'tasks')
PIPELINES_DIR = File.join(CONCOURSE_DIR, 'pipelines')

def concourse_tasks
  Dir
    .glob(File.join(TASKS_DIR, '*.yml'))
    .map { |f| [File.basename(f), File.read(f)] }
end

def concourse_pipelines
  Dir
    .glob(File.join(PIPELINES_DIR, '*.yml'))
    .map { |f| [File.basename(f), File.read(f)] }
end

def all_image_resources(frag)
  if frag.is_a?(Array)
    frag.flat_map { |val| all_image_resources(val) }
  elsif !frag.is_a?(Hash)
    []
  elsif [frag.dig('source', 'repository'), frag.dig('source', 'tag')].none?
    frag.values.flat_map { |val| all_image_resources(val) }
  else
    [{ repository: frag.dig('source', 'repository'),
       tag: frag.dig('source', 'tag') }]
  end
end

RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  # These two settings work together to allow you to limit a spec run
  # to individual examples or groups you care about by tagging them with
  # `:focus` metadata. When nothing is tagged with `:focus`, all examples
  # get run.
  config.filter_run :focus
  config.run_all_when_everything_filtered = true

  # Allows RSpec to persist some state between runs in order to support
  # the `--only-failures` and `--next-failure` CLI options. We recommend
  # you configure your source control system to ignore this file.
  config.example_status_persistence_file_path = "spec/examples.txt"

  # Limits the available syntax to the non-monkey patched syntax that is
  # recommended. For more details, see:
  #   - http://rspec.info/blog/2012/06/rspecs-new-expectation-syntax/
  #   - http://www.teaisaweso.me/blog/2013/05/27/rspecs-new-message-expectation-syntax/
  #   - http://rspec.info/blog/2014/05/notable-changes-in-rspec-3/#zero-monkey-patching-mode
  config.disable_monkey_patching!

  # This setting enables warnings. It's recommended, but in some cases may
  # be too noisy due to issues in dependencies.
  #config.warnings = true

  # Many RSpec users commonly either run the entire suite or an individual
  # file, and it's useful to allow more verbose output when running an
  # individual spec file.
  if config.files_to_run.one?
    # Use the documentation formatter for detailed output,
    # unless a formatter has already been configured
    # (e.g. via a command-line flag).
    config.default_formatter = 'doc'
  end

  # Print the 10 slowest examples and example groups at the
  # end of the spec run, to help surface which specs are running
  # particularly slow.
  #config.profile_examples = 10

  # Run specs in random order to surface order dependencies. If you find an
  # order dependency and want to debug it, you can fix the order by providing
  # the seed, which is printed after each run.
  #     --seed 1234
  config.order = :random

  # Seed global randomization in this process using the `--seed` CLI option.
  # Setting this allows you to use `--seed` to deterministically reproduce
  # test failures related to randomization by passing the same `--seed` value
  # as the one that triggered the failure.
  Kernel.srand config.seed
end

require 'fluent/plugin/filter'

module Fluent::Plugin
  class PassthruFilter < Fluent::Plugin::Filter
    # Register this filter as "passthru"
    Fluent::Plugin.register_filter('passthru', self)

    def filter(tag, time, record)
      record
    end
  end
end